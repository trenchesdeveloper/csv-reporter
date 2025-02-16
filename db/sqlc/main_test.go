package db

import (
	"database/sql"
	"errors"
	"fmt"
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
	cfg, err := config.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	connPool, err := sql.Open(cfg.DBDRIVER, cfg.DB_SOURCE_TEST)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("cannot get working directory:", err)
	}

	projectRoot := path.Join(wd, "..")
	migrationDir := path.Join(projectRoot, "migrations")
	migrationPath := fmt.Sprintf("file:///%s", migrationDir)
	log.Println("Migration path:", migrationPath)

	migration, err := migrate.New(migrationPath, cfg.DB_SOURCE_TEST)
	if err != nil {
		log.Fatal("failed to create migration instance:", err)
	}

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {

		log.Println("postgres migrated")

		defer connPool.Close()

		testStore = NewStore(connPool)
		os.Exit(m.Run())
	}
}
