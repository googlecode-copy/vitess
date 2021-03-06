// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tabletconntest provides the test methods to make sure a
// tabletconn/queryservice pair over RPC works correctly.
package tabletconntest

import (
	"reflect"
	"testing"

	mproto "github.com/youtube/vitess/go/mysql/proto"
	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/tabletserver/proto"
	"github.com/youtube/vitess/go/vt/tabletserver/queryservice"
	"github.com/youtube/vitess/go/vt/tabletserver/tabletconn"
	"golang.org/x/net/context"
)

// fakeQueryService has the server side of this fake
type fakeQueryService struct {
	t *testing.T
}

// TestKeyspace is the Keyspace we use for this test
const TestKeyspace = "test_keyspace"

// TestShard is the Shard we use for this test
const TestShard = "test_shard"

const testSessionId int64 = 5678

// GetSessionId is part of the queryservice.QueryService interface
func (f *fakeQueryService) GetSessionId(sessionParams *proto.SessionParams, sessionInfo *proto.SessionInfo) error {
	if sessionParams.Keyspace != TestKeyspace {
		f.t.Errorf("invalid keyspace: got %v expected %v", sessionParams.Keyspace, TestKeyspace)
	}
	if sessionParams.Shard != TestShard {
		f.t.Errorf("invalid shard: got %v expected %v", sessionParams.Shard, TestShard)
	}
	sessionInfo.SessionId = testSessionId
	return nil
}

// Begin is part of the queryservice.QueryService interface
func (f *fakeQueryService) Begin(ctx context.Context, session *proto.Session, txInfo *proto.TransactionInfo) error {
	if session.SessionId != testSessionId {
		f.t.Errorf("Begin: invalid SessionId: got %v expected %v", session.SessionId, testSessionId)
	}
	if session.TransactionId != 0 {
		f.t.Errorf("Begin: invalid TransactionId: got %v expected 0", session.TransactionId)
	}
	txInfo.TransactionId = beginTransactionId
	return nil
}

const beginTransactionId int64 = 9990

func testBegin(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testBegin")
	ctx := context.Background()
	transactionId, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if transactionId != beginTransactionId {
		t.Errorf("Unexpected result from Begin: got %v wanted %v", transactionId, beginTransactionId)
	}
}

// Commit is part of the queryservice.QueryService interface
func (f *fakeQueryService) Commit(ctx context.Context, session *proto.Session) error {
	if session.SessionId != testSessionId {
		f.t.Errorf("Commit: invalid SessionId: got %v expected %v", session.SessionId, testSessionId)
	}
	if session.TransactionId != commitTransactionId {
		f.t.Errorf("Commit: invalid TransactionId: got %v expected %v", session.TransactionId, commitTransactionId)
	}
	return nil
}

const commitTransactionId int64 = 999044

func testCommit(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testCommit")
	ctx := context.Background()
	err := conn.Commit(ctx, commitTransactionId)
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
}

// Rollback is part of the queryservice.QueryService interface
func (f *fakeQueryService) Rollback(ctx context.Context, session *proto.Session) error {
	if session.SessionId != testSessionId {
		f.t.Errorf("Rollback: invalid SessionId: got %v expected %v", session.SessionId, testSessionId)
	}
	if session.TransactionId != rollbackTransactionId {
		f.t.Errorf("Rollback: invalid TransactionId: got %v expected %v", session.TransactionId, rollbackTransactionId)
	}
	return nil
}

const rollbackTransactionId int64 = 999044

func testRollback(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testRollback")
	ctx := context.Background()
	err := conn.Rollback(ctx, rollbackTransactionId)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

// Execute is part of the queryservice.QueryService interface
func (f *fakeQueryService) Execute(ctx context.Context, query *proto.Query, reply *mproto.QueryResult) error {
	if query.Sql != executeQuery {
		f.t.Errorf("invalid Execute.Query.Sql: got %v expected %v", query.Sql, executeQuery)
	}
	if !reflect.DeepEqual(query.BindVariables, executeBindVars) {
		f.t.Errorf("invalid Execute.Query.BindVariables: got %v expected %v", query.BindVariables, executeBindVars)
	}
	if query.SessionId != testSessionId {
		f.t.Errorf("invalid Execute.Query.SessionId: got %v expected %v", query.SessionId, testSessionId)
	}
	if query.TransactionId != executeTransactionId {
		f.t.Errorf("invalid Execute.Query.TransactionId: got %v expected %v", query.TransactionId, executeTransactionId)
	}
	*reply = executeQueryResult
	return nil
}

const executeQuery = "executeQuery"

var executeBindVars = map[string]interface{}{
	"bind1": int64(1114444),
}

const executeTransactionId int64 = 678

var executeQueryResult = mproto.QueryResult{
	Fields: []mproto.Field{
		mproto.Field{
			Name: "field1",
			Type: 42,
		},
		mproto.Field{
			Name: "field2",
			Type: 73,
		},
	},
	RowsAffected: 123,
	InsertId:     72,
	Rows: [][]sqltypes.Value{
		[]sqltypes.Value{
			sqltypes.MakeString([]byte("row1 value1")),
			sqltypes.MakeString([]byte("row1 value2")),
		},
		[]sqltypes.Value{
			sqltypes.MakeString([]byte("row2 value1")),
			sqltypes.MakeString([]byte("row2 value2")),
		},
	},
}

func testExecute(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testExecute")
	ctx := context.Background()
	qr, err := conn.Execute(ctx, executeQuery, executeBindVars, executeTransactionId)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !reflect.DeepEqual(*qr, executeQueryResult) {
		t.Errorf("Unexpected result from Execute: got %v wanted %v", qr, executeQueryResult)
	}
}

// StreamExecute is part of the queryservice.QueryService interface
func (f *fakeQueryService) StreamExecute(ctx context.Context, query *proto.Query, sendReply func(*mproto.QueryResult) error) error {
	if query.Sql != streamExecuteQuery {
		f.t.Errorf("invalid StreamExecute.Query.Sql: got %v expected %v", query.Sql, streamExecuteQuery)
	}
	if !reflect.DeepEqual(query.BindVariables, streamExecuteBindVars) {
		f.t.Errorf("invalid StreamExecute.Query.BindVariables: got %v expected %v", query.BindVariables, streamExecuteBindVars)
	}
	if query.SessionId != testSessionId {
		f.t.Errorf("invalid StreamExecute.Query.SessionId: got %v expected %v", query.SessionId, testSessionId)
	}
	if query.TransactionId != streamExecuteTransactionId {
		f.t.Errorf("invalid StreamExecute.Query.TransactionId: got %v expected %v", query.TransactionId, streamExecuteTransactionId)
	}
	if err := sendReply(&streamExecuteQueryResult1); err != nil {
		f.t.Errorf("sendReply1 failed: %v", err)
	}
	if err := sendReply(&streamExecuteQueryResult2); err != nil {
		f.t.Errorf("sendReply2 failed: %v", err)
	}
	return nil
}

const streamExecuteQuery = "streamExecuteQuery"

var streamExecuteBindVars = map[string]interface{}{
	"bind1": int64(93848000),
}

const streamExecuteTransactionId int64 = 6789992

var streamExecuteQueryResult1 = mproto.QueryResult{
	Fields: []mproto.Field{
		mproto.Field{
			Name: "field1",
			Type: 42,
		},
		mproto.Field{
			Name: "field2",
			Type: 73,
		},
	},
}

var streamExecuteQueryResult2 = mproto.QueryResult{
	Rows: [][]sqltypes.Value{
		[]sqltypes.Value{
			sqltypes.MakeString([]byte("row1 value1")),
			sqltypes.MakeString([]byte("row1 value2")),
		},
		[]sqltypes.Value{
			sqltypes.MakeString([]byte("row2 value1")),
			sqltypes.MakeString([]byte("row2 value2")),
		},
	},
}

func testStreamExecute(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testStreamExecute")
	ctx := context.Background()
	stream, errFunc, err := conn.StreamExecute(ctx, streamExecuteQuery, streamExecuteBindVars, streamExecuteTransactionId)
	if err != nil {
		t.Fatalf("StreamExecute failed: %v", err)
	}
	qr, ok := <-stream
	if !ok {
		t.Fatalf("StreamExecute failed: cannot read result1")
	}
	if len(qr.Rows) == 0 {
		qr.Rows = nil
	}
	if !reflect.DeepEqual(*qr, streamExecuteQueryResult1) {
		t.Errorf("Unexpected result1 from StreamExecute: got %v wanted %v", qr, streamExecuteQueryResult1)
	}
	qr, ok = <-stream
	if !ok {
		t.Fatalf("StreamExecute failed: cannot read result2")
	}
	if len(qr.Fields) == 0 {
		qr.Fields = nil
	}
	if !reflect.DeepEqual(*qr, streamExecuteQueryResult2) {
		t.Errorf("Unexpected result2 from StreamExecute: got %v wanted %v", qr, streamExecuteQueryResult2)
	}
	qr, ok = <-stream
	if ok {
		t.Fatalf("StreamExecute channel wasn't closed")
	}
	if err := errFunc(); err != nil {
		t.Fatalf("StreamExecute errFunc failed: %v", err)
	}
}

// ExecuteBatch is part of the queryservice.QueryService interface
func (f *fakeQueryService) ExecuteBatch(ctx context.Context, queryList *proto.QueryList, reply *proto.QueryResultList) error {
	if !reflect.DeepEqual(queryList.Queries, executeBatchQueries) {
		f.t.Errorf("invalid ExecuteBatch.QueryList.Queries: got %v expected %v", queryList.Queries, executeBatchQueries)
	}
	if queryList.SessionId != testSessionId {
		f.t.Errorf("invalid ExecuteBatch.QueryList.SessionId: got %v expected %v", queryList.SessionId, testSessionId)
	}
	if queryList.TransactionId != executeBatchTransactionId {
		f.t.Errorf("invalid ExecuteBatch.QueryList.TransactionId: got %v expected %v", queryList.TransactionId, executeBatchTransactionId)
	}
	*reply = executeBatchQueryResultList
	return nil
}

var executeBatchQueries = []proto.BoundQuery{
	proto.BoundQuery{
		Sql: "executeBatchQueries1",
		BindVariables: map[string]interface{}{
			"bind1": int64(43),
		},
	},
	proto.BoundQuery{
		Sql: "executeBatchQueries2",
		BindVariables: map[string]interface{}{
			"bind2": int64(72),
		},
	},
}

const executeBatchTransactionId int64 = 678

var executeBatchQueryResultList = proto.QueryResultList{
	List: []mproto.QueryResult{
		mproto.QueryResult{
			Fields: []mproto.Field{
				mproto.Field{
					Name: "field1",
					Type: 46,
				},
			},
			RowsAffected: 1232,
			InsertId:     712,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("row1 value1")),
				},
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("row2 value1")),
				},
			},
		},
		mproto.QueryResult{
			Fields: []mproto.Field{
				mproto.Field{
					Name: "field1",
					Type: 42,
				},
			},
			RowsAffected: 12333,
			InsertId:     74442,
			Rows: [][]sqltypes.Value{
				[]sqltypes.Value{
					sqltypes.MakeString([]byte("row1 value1")),
					sqltypes.MakeString([]byte("row1 value2")),
				},
			},
		},
	},
}

func testExecuteBatch(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testExecuteBatch")
	ctx := context.Background()
	qrl, err := conn.ExecuteBatch(ctx, executeBatchQueries, executeBatchTransactionId)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}
	if !reflect.DeepEqual(*qrl, executeBatchQueryResultList) {
		t.Errorf("Unexpected result from Execute: got %v wanted %v", qrl, executeBatchQueryResultList)
	}
}

// SplitQuery is part of the queryservice.QueryService interface
func (f *fakeQueryService) SplitQuery(ctx context.Context, req *proto.SplitQueryRequest, reply *proto.SplitQueryResult) error {
	if !reflect.DeepEqual(req.Query, splitQueryBoundQuery) {
		f.t.Errorf("invalid SplitQuery.SplitQueryRequest.Query: got %v expected %v", req.Query, splitQueryBoundQuery)
	}
	if req.SplitCount != splitQuerySplitCount {
		f.t.Errorf("invalid SplitQuery.SplitQueryRequest.SplitCount: got %v expected %v", req.SplitCount, splitQuerySplitCount)
	}
	reply.Queries = splitQueryQuerySplitList
	return nil
}

var splitQueryBoundQuery = proto.BoundQuery{
	Sql: "splitQuery",
	BindVariables: map[string]interface{}{
		"bind1": int64(43),
	},
}

const splitQuerySplitCount = 372

var splitQueryQuerySplitList = []proto.QuerySplit{
	proto.QuerySplit{
		Query: proto.BoundQuery{
			Sql: "splitQuery",
			BindVariables: map[string]interface{}{
				"bind1":       int64(43),
				"keyspace_id": int64(3333),
			},
		},
		RowCount: 4456,
	},
}

func testSplitQuery(t *testing.T, conn tabletconn.TabletConn) {
	t.Log("testSplitQuery")
	ctx := context.Background()
	qsl, err := conn.SplitQuery(ctx, splitQueryBoundQuery, splitQuerySplitCount)
	if err != nil {
		t.Fatalf("SplitQuery failed: %v", err)
	}
	if !reflect.DeepEqual(qsl, splitQueryQuerySplitList) {
		t.Errorf("Unexpected result from SplitQuery: got %v wanted %v", qsl, splitQueryQuerySplitList)
	}
}

// CreateFakeServer returns the fake server for the tests
func CreateFakeServer(t *testing.T) queryservice.QueryService {
	return &fakeQueryService{t}
}

// TestSuite runs all the tests
func TestSuite(t *testing.T, conn tabletconn.TabletConn) {
	testBegin(t, conn)
	testCommit(t, conn)
	testRollback(t, conn)
	testExecute(t, conn)
	testStreamExecute(t, conn)
	testExecuteBatch(t, conn)
	testSplitQuery(t, conn)
}
