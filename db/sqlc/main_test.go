package db

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path"
	"testing"

	"github.com/trenchesdeveloper/csv-reporter/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var testStore Store

func TestMain(m *testing.M) {
	// 1) Load config & open DB
	cfg, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	connPool, err := sql.Open(cfg.DBDRIVER, cfg.DB_SOURCE_TEST)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	// 2) Build migration path
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("cannot get working directory: %v", err)
	}
	projectRoot := path.Join(wd, "..")
	migrationDir := path.Join(projectRoot, "migrations")
	// Use exactly three slashes then the absolute path
	migrationPath := "file:///" + migrationDir
	log.Println("Migration path:", migrationPath)

	// 3) Run migrations (ignore ErrNoChange)
	migr, err := migrate.New(migrationPath, cfg.DB_SOURCE_TEST)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := migr.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("failed to run up migrations: %v", err)
	}

	// 4) Init store & run tests
	testStore = NewStore(connPool)
	code := m.Run()

	// 5) Cleanup & exit
	connPool.Close()
	os.Exit(code)
}
