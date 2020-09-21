package main

import (
	"context"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type config struct {
	dbDSN     string
	direction string
	fixtures  bool
}

func main() {
	var c config
	flag.StringVar(&c.dbDSN, "db-dsn", "postgres://root:root@localhost:5432/paybile", "Database connection DSN.")
	flag.StringVar(&c.direction, "direction", "up", "Migrations direction: up, down.")
	flag.BoolVar(&c.fixtures, "fixtures", false, "Reload fictures: true, false.")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	trapSignals(cancel)

	db, closeDB := connectDB(ctx, c.dbDSN)
	defer closeDB()

	d, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("cannot instantiate postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "", d)
	if err != nil {
		log.Fatalf("cannot instantiate migrator for schema: %v", err)
	}
	if c.direction == "up" {
		if err = m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrations up failed: %v", err)
		}
	}
	if c.direction == "down" {
		if err = m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrations down failed: %v", err)
		}
	}

	if !c.fixtures {
		return
	}

	query, err := ioutil.ReadFile("fixtures/db_fixtures.sql")
	if err != nil {
		log.Fatalf("cannot read fixtures file: %v", err)
	}
	if _, err = db.ExecContext(ctx, string(query)); err != nil {
		log.Fatalf("cannot execute fixtures: %v", err)
	}
}

func trapSignals(cancel func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()
}

// connectDB tries to establish a connection
// to the database until the context is
// so the DB can initialise.
func connectDB(ctx context.Context, dsn string) (*sql.DB, func()) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Fatalf("cannot open connection to db: %v", ctx.Err())
		case <-t.C:
			db, err := sql.Open("postgres", dsn)
			if err != nil {
				continue
			}
			if err = db.PingContext(ctx); err != nil {
				continue
			}

			return db, func() {
				_ = db.Close()
			}
		}
	}
}

