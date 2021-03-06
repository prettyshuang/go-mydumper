/*
 * go-mydumper
 * xelabs.org
 *
 * Copyright (c) XeLabs
 * GPL License
 *
 */

package common

import (
	"testing"

	"github.com/XeLabs/go-mysqlstack/driver"
	querypb "github.com/XeLabs/go-mysqlstack/sqlparser/depends/query"
	"github.com/XeLabs/go-mysqlstack/sqlparser/depends/sqltypes"
	"github.com/XeLabs/go-mysqlstack/xlog"
	"github.com/stretchr/testify/assert"
)

func TestStreamer(t *testing.T) {
	log := xlog.NewStdLog(xlog.Level(xlog.INFO))
	fromFakedbs := driver.NewTestHandler(log)
	toFakedbs := driver.NewTestHandler(log)

	fromSvr, err := driver.MockMysqlServer(log, fromFakedbs)
	assert.Nil(t, err)
	defer fromSvr.Close()
	fromAddr := fromSvr.Addr()

	toSvr, err := driver.MockMysqlServer(log, toFakedbs)
	assert.Nil(t, err)
	defer toSvr.Close()
	toAddr := toSvr.Addr()

	selectResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "id",
				Type: querypb.Type_INT32,
			},
			{
				Name: "name",
				Type: querypb.Type_VARCHAR,
			},
			{
				Name: "namei1",
				Type: querypb.Type_VARCHAR,
			},
			{
				Name: "null",
				Type: querypb.Type_NULL_TYPE,
			},
			{
				Name: "decimal",
				Type: querypb.Type_DECIMAL,
			},
			{
				Name: "datetime",
				Type: querypb.Type_DATETIME,
			},
		},
		Rows: make([][]sqltypes.Value, 0, 256)}

	for i := 0; i < 201710; i++ {
		row := []sqltypes.Value{
			sqltypes.MakeTrusted(querypb.Type_INT32, []byte("11")),
			sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("11\"xx\"")),
			sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("")),
			sqltypes.MakeTrusted(querypb.Type_NULL_TYPE, nil),
			sqltypes.MakeTrusted(querypb.Type_DECIMAL, []byte("210.01")),
			sqltypes.NULL,
		}
		selectResult.Rows = append(selectResult.Rows, row)
	}

	schemaResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "Table",
				Type: querypb.Type_VARCHAR,
			},
			{
				Name: "Create Table",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t1")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("CREATE TABLE `t1` (`a` int(11) DEFAULT NULL,`b` varchar(100) DEFAULT NULL) ENGINE=InnoDB")),
			},
		}}

	tablesResult := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "Tables_in_test",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t1")),
			},
			{
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("t2")),
			},
		}}

	// fakedbs.
	{
		fromFakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		fromFakedbs.AddQueryPattern("show create table .*", schemaResult)
		fromFakedbs.AddQueryPattern("show tables from .*", tablesResult)
		fromFakedbs.AddQueryPattern("select .*", selectResult)

		toFakedbs.AddQueryPattern("use .*", &sqltypes.Result{})
		toFakedbs.AddQueryPattern("create database .*", &sqltypes.Result{})
		toFakedbs.AddQueryPattern("create table .*", &sqltypes.Result{})
		toFakedbs.AddQueryPattern("insert into .*", &sqltypes.Result{})
		toFakedbs.AddQueryPattern("drop table .*", &sqltypes.Result{})
	}

	args := &Args{
		Database:        "test",
		User:            "mock",
		Password:        "mock",
		Address:         fromAddr,
		ToUser:          "mock",
		ToPassword:      "mock",
		ToAddress:       toAddr,
		ChunksizeInMB:   1,
		Threads:         16,
		StmtSize:        10000,
		IntervalMs:      500,
		OverwriteTables: true,
	}

	// Streamer.
	{
		Streamer(log, args)
	}
}
