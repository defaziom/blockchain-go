package db

import "github.com/hashicorp/go-memdb"

func GetSchema() *memdb.DBSchema {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"peer": &memdb.TableSchema{
				Name: "peer",
				Indexes: map[string]*memdb.IndexSchema{
					"ip": &memdb.IndexSchema{
						Name:    "ip",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Ip"},
					},
					"port": &memdb.IndexSchema{
						Name:    "port",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "Port"},
					},
				},
			},
		},
	}
	return schema
}
