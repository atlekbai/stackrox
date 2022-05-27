// Code generated by pg-bindings generator. DO NOT EDIT.

package schema

import (
	"reflect"
	"time"

	"github.com/lib/pq"
	"github.com/stackrox/rox/central/globaldb"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/postgres"
	"github.com/stackrox/rox/pkg/postgres/walker"
	"github.com/stackrox/rox/pkg/search"
)

var (
	// CreateTableTestMultiKeyStructsStmt holds the create statement for table `test_multi_key_structs`.
	CreateTableTestMultiKeyStructsStmt = &postgres.CreateStmts{
		Table: `
               create table if not exists test_multi_key_structs (
                   Key1 varchar,
                   Key2 varchar,
                   StringSlice text[],
                   Bool bool,
                   Uint64 integer,
                   Int64 integer,
                   Float numeric,
                   Labels jsonb,
                   Timestamp timestamp,
                   Enum integer,
                   Enums int[],
                   String_ varchar,
                   IntSlice int[],
                   Oneofnested_Nested varchar,
                   serialized bytea,
                   PRIMARY KEY(Key1, Key2)
               )
               `,
		GormModel: (*TestMultiKeyStructs)(nil),
		Indexes:   []string{},
		Children: []*postgres.CreateStmts{
			&postgres.CreateStmts{
				Table: `
               create table if not exists test_multi_key_structs_nesteds (
                   test_multi_key_structs_Key1 varchar,
                   test_multi_key_structs_Key2 varchar,
                   idx integer,
                   Nested varchar,
                   IsNested bool,
                   Int64 integer,
                   Nested2_Nested2 varchar,
                   Nested2_IsNested bool,
                   Nested2_Int64 integer,
                   PRIMARY KEY(test_multi_key_structs_Key1, test_multi_key_structs_Key2, idx),
                   CONSTRAINT fk_parent_table_0 FOREIGN KEY (test_multi_key_structs_Key1, test_multi_key_structs_Key2) REFERENCES test_multi_key_structs(Key1, Key2) ON DELETE CASCADE
               )
               `,
				GormModel: (*TestMultiKeyStructsNesteds)(nil),
				Indexes: []string{
					"create index if not exists testMultiKeyStructsNesteds_idx on test_multi_key_structs_nesteds using btree(idx)",
				},
				Children: []*postgres.CreateStmts{},
			},
		},
	}

	// TestMultiKeyStructsSchema is the go schema for table `test_multi_key_structs`.
	TestMultiKeyStructsSchema = func() *walker.Schema {
		schema := globaldb.GetSchemaForTable("test_multi_key_structs")
		if schema != nil {
			return schema
		}
		schema = walker.Walk(reflect.TypeOf((*storage.TestMultiKeyStruct)(nil)), "test_multi_key_structs")
		schema.SetOptionsMap(search.Walk(v1.SearchCategory_SEARCH_UNSET, "testmultikeystruct", (*storage.TestMultiKeyStruct)(nil)))
		globaldb.RegisterTable(schema)
		return schema
	}()
)

const (
	TestMultiKeyStructsTableName        = "test_multi_key_structs"
	TestMultiKeyStructsNestedsTableName = "test_multi_key_structs_nesteds"
)

// TestMultiKeyStructs holds the Gorm model for Postgres table `test_multi_key_structs`.
type TestMultiKeyStructs struct {
	Key1              string                          `gorm:"column:key1;type:varchar;primaryKey"`
	Key2              string                          `gorm:"column:key2;type:varchar;primaryKey"`
	StringSlice       *pq.StringArray                 `gorm:"column:stringslice;type:text[]"`
	Bool              bool                            `gorm:"column:bool;type:bool"`
	Uint64            uint64                          `gorm:"column:uint64;type:integer"`
	Int64             int64                           `gorm:"column:int64;type:integer"`
	Float             float32                         `gorm:"column:float;type:numeric"`
	Labels            map[string]string               `gorm:"column:labels;type:jsonb"`
	Timestamp         *time.Time                      `gorm:"column:timestamp;type:timestamp"`
	Enum              storage.TestMultiKeyStruct_Enum `gorm:"column:enum;type:integer"`
	Enums             *pq.Int32Array                  `gorm:"column:enums;type:int[]"`
	String            string                          `gorm:"column:string_;type:varchar"`
	IntSlice          *pq.Int32Array                  `gorm:"column:intslice;type:int[]"`
	OneofnestedNested string                          `gorm:"column:oneofnested_nested;type:varchar"`
	Serialized        []byte                          `gorm:"column:serialized;type:bytea"`
}

// TestMultiKeyStructsNesteds holds the Gorm model for Postgres table `test_multi_key_structs_nesteds`.
type TestMultiKeyStructsNesteds struct {
	TestMultiKeyStructsKey1 string              `gorm:"column:test_multi_key_structs_key1;type:varchar;primaryKey"`
	TestMultiKeyStructsKey2 string              `gorm:"column:test_multi_key_structs_key2;type:varchar;primaryKey"`
	Idx                     int                 `gorm:"column:idx;type:integer;primaryKey;index:testmultikeystructsnesteds_idx,type:btree"`
	Nested                  string              `gorm:"column:nested;type:varchar"`
	IsNested                bool                `gorm:"column:isnested;type:bool"`
	Int64                   int64               `gorm:"column:int64;type:integer"`
	Nested2Nested2          string              `gorm:"column:nested2_nested2;type:varchar"`
	Nested2IsNested         bool                `gorm:"column:nested2_isnested;type:bool"`
	Nested2Int64            int64               `gorm:"column:nested2_int64;type:integer"`
	TestMultiKeyStructsRef  TestMultiKeyStructs `gorm:"foreignKey:test_multi_key_structs_key1,test_multi_key_structs_key2;references:key1,key2;belongsTo;constraint:OnDelete:CASCADE"`
}