package main

import (
	"GoFlow_/config"
	"database/sql"
	_ "github.com/lib/pq"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

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
