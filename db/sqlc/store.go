package db

import "database/sql"

type Store interface {
	Querier
}

type SQLStore struct {
	connPool *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		connPool: db,
		Queries:  New(db),
	}
}
