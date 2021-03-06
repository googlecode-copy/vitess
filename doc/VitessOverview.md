Vitess is a database solution for scaling MySQL. It's architected to run as
effectively in a public or private cloud architecture as it does on dedicated
hardware. It combines and extends many important MySQL features with the
scalability of a NoSQL database. Vitess has been serving all YouTube database
traffic since 2011.

### Vitess on Kubernetes

Kubernetes is an open-source orchestration system for Docker containers, and Vitess is the logical storage engine choice for Kubernetes users.

Kubernetes handles scheduling onto nodes in a compute cluster, actively manages workloads on those nodes, and groups containers comprising an application for easy management and discovery. Using Kubernetes, you can easily create and manage a Vitess cluster, out of the box.

## Comparisons to other storage options

The following sections compare Vitess to two common alternatives, a vanilla MySQL implementation and a NoSQL implementation.

### Vitess vs. Vanilla MySQL

Vitess improves a vanilla MySQL implementation in several ways:

<table class="comparison">
  <tr>
    <th>Vanilla MySQL</th>
    <th>Vitess</th>
  </tr>
  <tr>
    <td>Every MySQL connection has a memory overhead that ranges between 256KB and almost 3MB, depending on which MySQL release you're using. As your user base grows, you need to add RAM to support additional connections, but the RAM does not contribute to faster queries. In addition, there is a significant CPU cost associated with obtaining the connections.</td>
    <td>Vitess' BSON-based protocol creates very lightweight connections that are around 32KB. Vitess' connection pooling feature uses Go's concurrency support to map these lightweight connections to a small pool of MySQL connections. As such, Vitess can easily handle thousands of connections.</td>
  </tr>
  <tr>
    <td>Poorly written queries, such as those that don't set a LIMIT, can negatively impact database performance for all users.</td>
    <td>Vitess employs a SQL parser that uses a configurable set of rules to rewrite queries that might hurt database performance.</td>
  </tr>
  <tr>
    <td>Sharding is a process of partitioning your data to improve scalability and performance. MySQL lacks native sharding support, requiring you to write sharding code and embed sharding logic in your application.</td>
    <td>Vitess uses range-based sharding. It supports both horizontal and vertical resharding, completing most data transitions with just a few seconds of read-only downtime. Vitess can even accommodate a custom sharding scheme that you already have in place.</td>
  </tr>
  <tr>
    <td>A MySQL cluster using replication for availability has a master database and a few replicas. If the master fails, a replica should become the new master. This requires you to manage the database lifecycle and communicate the current system state to your application.</td>
    <td>Vitess helps to manage the lifecycle of your database scenarios. It supports and automatically handles various scenarios, including master failover and data backups.</td>
  </tr>
  <tr>
    <td>A MySQL cluster can have custom database configurations for different workloads, like a master database for writes, fast read-only replicas for web clients, slower read-only replicas for batch jobs, and so forth. If the database has horizontal sharding, the setup is repeated for each shard, and the app needs baked-in logic to know how to find the right database.</td>
    <td>Vitess uses a topology backed by a consistent data store, like etcd or ZooKeeper. This means the cluster view is always up-to-date and consistent for different clients. Vitess also provides a proxy that routes queries efficiently to the most appropriate MySQL instance.</td>
  </tr>
</table>

### Vitess vs. NoSQL

If you're considering a NoSQL solution primarily because of concerns about the scalability of MySQL, Vitess might be a more appropriate choice for your application. While NoSQL provides great support for unstructured data, Vitess still offers several benefits not available in NoSQL datastores:

<table class="comparison">
  <tr>
    <th>NoSQL</th>
    <th>Vitess</th>
  </tr>
  <tr>
    <td>NoSQL databases do not define relationships between database tables, and only support a subset of the SQL language.</td>
    <td>Vitess is not a simple key-value store. It supports complex query semantics such as where clauses, JOINS, aggregation functions, and more.</td>
  </tr>
  <tr>
    <td>NoSQL datastores do not support transactions.</td>
    <td>Vitess supports transactions within a shard. We are also exploring the feasibility of supporting cross-shard transactions using 2PC.</td>
  </tr>
  <tr>
    <td>NoSQL solutions have custom APIs, leading to custom architectures, applications, and tools.</td>
    <td>Vitess adds very little variance to MySQL, a database that most people are already accustomed to working with.</td>
  </tr>
  <tr>
    <td>NoSQL solutions provide limited support for database indexes compared to MySQL.</td>
    <td>Vitess allows you to use all of MySQL's indexing functionality to optimize query performance.</td>
  </tr>
</table>

## Features

* **Performance**
  * Connection pooling - Scale front-end connections while optimizing MySQL performance.
  * Query de-duping – Reuse results of an in-flight query for any identical requests received while the in-flight query was still executing.
  * Transaction manager – Limit number of concurrent transactions and manage deadlines to optimize overall throughput.
  * Rowcache – Maintain a row-based cache (using memcached) to more efficiently field queries that require random access by primary key, very useful for OLTP workloads. (The MySQL buffer cache is optimized for range scans over indices and tables.). This can replace a custom caching layer implementation at the application layer.<br><br>

* **Protection**

  * Query rewriting and sanitation – Add limits and avoid non-deterministic updates.
  * Query blacklisting – Customize rules to prevent potentially problematic queries from hitting your database.
  * Query killer – Terminate queries that take too long to return data.
  * Table ACLs – Specify access control lists (ACLs) for tables based on the connected user.

* **Monitoring**
  * Performance analysis: Tools let you monitor, diagnose, and analyze your database performance.
  * Query streaming – Use a list of incoming queries to serve OLAP workloads.
  * Update stream – A server streams the list of rows changing in the database, which can be used as a mechanism to propagate changes to other data stores.

* **Topology Management Tools**
  * Master management tools (handles reparenting)
  * Web-based management GUI
  * Designed to work in multiple data centers / regions

* **Sharding**
  * Virtually seamless dynamic re-sharding
  * Vertical and Horizontal sharding support
  * Built-in range-based, or application-defined sharding support


## Architecture

The Vitess platform consists of a number of server processes, command-line utilities, and web-based utilities, backed by a consistent metadata store.

Depending on the current state of your application, you could arrive at a full Vitess implementation through a number of different process flows. For example, if you're building a service from scratch, your first step with Vitess would be to define your database topology. However, if you need to scale your existing database, you'd likely start by deploying a connection proxy.

The diagram below illustrates Vitess' components:

![Diagram showing Vitess implementation](https://raw.githubusercontent.com/youtube/vitess/master/doc/VitessOverview.png)

### Topology

The [Topology Service](https://github.com/youtube/vitess/blob/master/doc/TopologyService.md) is a metadata store that contains information about running servers, the sharding scheme, and the replication graph.  The topology is backed by a consistent data store.  You can explore the topology using **vtctl** (command-line) and **vtctld** (web).

In Kubernetes, the data store is [etcd](https://github.com/coreos/etcd).  Vitess source code also ships with [Apache ZooKeeper](http://zookeeper.apache.org/) support.

### vttablet

**vttablet** is a server that sits in front of a MySQL database. It is a newer version of and provides all of the same benefits as vtocc, including connection pooling, query rewriting, and query de-duping. In addition, vttablet executes management tasks that vtctl initiates. It also provides streaming services that are used for [filtered replication](Resharding.md#filtered-replication) and data export.

A Vitess implementation has one vttablet for each MySQL instance.

A lightweight Vitess implementation uses a smart connection proxy that serves queries for a single MySQL database. Running the proxy server in front of your MySQL database and changing your app to use the Vitess client instead of your MySQL driver provides Vitess' connection pooling feature.

### vtgate

**vtgate** is a light proxy server that routes traffic to the correct vttablet. To do so, vtgate considers the sharding scheme, required latency, and the availability of the tablets and their underlying MySQL instances. This allows the client to be very simple since it only needs to be able to find a vtgate instance.

### vtctl

**vtctl** is a command-line tool used to administer a Vitess cluster. It allows a human or application to easily interact with a Vitess implementation. Using vtctl, you can identify master and replica databases, create tables, initiate failovers, perform sharding (and resharding) operations, and so forth.
