package main

import (
	"context"
	"database/sql"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hmoragrega/paybile/repository/postgres"
	"github.com/hmoragrega/paybile/service"
	httptransport "github.com/hmoragrega/paybile/transport/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	port  string
	dbDSN string
}

func main() {
	var conf config
	flag.StringVar(&conf.port, "port", "8080", "API service port.")
	flag.StringVar(&conf.dbDSN, "db-dsn", "postgres://root:root@localhost:5432/paybile", "Database connection DSN.")
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	initCtx, initCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer initCancel()

	db, closeDB := connectDB(initCtx, conf.dbDSN)
	defer closeDB()

	var (
		// Repositories.
		userRepo        = &postgres.UserRepository{DB: db}
		walletRepo      = &postgres.WalletRepository{DB: db}
		transactionRepo = &postgres.TransactionRepository{DB: db}
		transferRepo    = &postgres.TransferRepository{
			DB:              db,
			WalletRepo:      walletRepo,
			TransactionRepo: transactionRepo,
		}
	)

	var (
		// Services.
		userSvc   = &service.UserService{Reader: userRepo}
		walletSvc = &service.WalletService{
			WalletReader:       walletRepo,
			TransferCreator:    transferRepo,
			TransactionsReader: transactionRepo,
		}
	)

	api := httptransport.ApiService{
		LoginService:      userSvc,
		TransactionLister: walletSvc,
		TransferCreator:   walletSvc,
		WalletGetter:      walletSvc,
	}

	s := http.Server{
		Addr:    ":" + conf.port,
		Handler: api.Handler(),
	}

	stop := make(chan struct{}, 2)
	trapSignals(stop)

	go func() {
		log.Info().Str("port", conf.port).Msg("Starting HTTP server")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server error")
			stop<- struct{}{}
		}
	}()

	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer shutdownCancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("error shutting down HTTP server")
	} else {
		log.Info().Msg("service shutdown completed")
	}
}

func trapSignals(stop chan<- struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		stop <- struct{}{}
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
			log.Fatal().Err(ctx.Err()).Msg("timeout connecting to db")
		case <-t.C:
			db, err := sql.Open("postgres", dsn)
			if err != nil {
				log.Warn().Err(err).Msg("cannot connect to db. retrying.")
				continue
			}
			if err = db.PingContext(ctx); err != nil {
				log.Warn().Err(err).Msg("cannot ping db. retrying.")
				continue
			}

			return db, func() {
				_ = db.Close()
			}
		}
	}
}
