package database

import "github.com/hashicorp/go-memdb"

var dbInstance *memdb.MemDB

func GetDatabase() *memdb.MemDB {
	if dbInstance == nil {
		db, err := memdb.NewMemDB(GetSchema())
		if err != nil {
			panic(err)
		}
		dbInstance = db
	}
	return dbInstance
}
