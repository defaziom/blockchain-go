package database

import "github.com/hashicorp/go-memdb"

func GetSchema() *memdb.DBSchema {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"peer_conn_info": &memdb.TableSchema{
				Name: "peer_conn_info",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Ip"},
					},
				},
			},
		},
	}
	return schema
}
