package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createTables = `
		CREATE TABLE IF NOT EXISTS checks
			(
				type              VARCHAR(64),
				name              VARCHAR(64),
				number_of_runs    INTEGER,
				number_of_failure INTEGER,
				last_updated      TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (type, name)
			);
	`
	createOrUpdateCheck = `
		INSERT INTO checks
			(type, name, number_of_runs, number_of_failure)
		VALUES
			(?, ?, 1, ?)
		ON CONFLICT
			(type, name)
		DO UPDATE SET
			number_of_runs=number_of_runs+1,
			number_of_failure=CASE WHEN ? > 0 THEN (number_of_failure + 1) ELSE 0 END,
			last_updated=CURRENT_TIMESTAMP;
	`
	getfailedChecks = `
		SELECT COUNT(*)
		FROM checks
		WHERE number_of_failure!=0
	  		AND last_updated >= Datetime('now', '-10 minutes');`
)

type Config struct {
	Database string `yaml:"database"`
}

type DB struct {
	db                  *sql.DB
	CreateOrUpdateCheck *sql.Stmt
	GetFailedChecks     *sql.Stmt
}

func Init(cfg Config) (DB, error) {

	db := DB{}
	var err error
	db.db, err = sql.Open("sqlite3", cfg.Database)
	if err != nil {
		return db, fmt.Errorf("unable to connect to database %v, %w", cfg.Database, err)
	}
	_, err = db.db.Exec(createTables)
	if err != nil {
		return db, fmt.Errorf("unable to create tables, %w", err)
	}
	log.Printf("Using database %v", cfg.Database)

	db.CreateOrUpdateCheck, err = db.db.Prepare(createOrUpdateCheck)
	if err != nil {
		return db, fmt.Errorf("failed to prepare statement, %w", err)
	}

	db.GetFailedChecks, err = db.db.Prepare(getfailedChecks)
	if err != nil {
		return db, fmt.Errorf("failed to prepare statement, %w", err)
	}

	return db, nil
}
