package main

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/trenchesdeveloper/csv-reporter/config"
)

func main() {
	cfg, err := config.LoadConfig(".")

	if err != nil {
		panic(err)
	}
	db, err := sql.Open(cfg.DBDRIVER, cfg.DBSOURCE)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Ping()

}
