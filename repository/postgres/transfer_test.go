//+build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

const (
	dbImage        = "postgres:9.6"
	migrationsPath = "../../migrations"
	fixturesFile   = "../../fixtures/db_fixtures.sql"
)

var (
	userA = service.User{ID: uuid.MustParse("bbc00191-b064-4655-9075-261ccef978cb")}

	walletA = uuid.MustParse("2f9b76dd-f689-456e-9080-6789718018a5")
	walletB = uuid.MustParse("4e1d841d-e53f-4785-ba4d-99df05f11eee")
	walletC = uuid.MustParse("f0212317-88db-4dd4-ba0e-39757e1ebcc6")
	walletD = uuid.MustParse("f889299f-41c4-4e58-96c2-7451c8276842")
)

func TestCreateTransfer(t *testing.T) {
	db := setupTestDB(t)
	tt := []struct {
		name                   string
		req                    service.TransferRequest
		checkBalance           bool
		wantTransfer           service.Transfer
		wantOriginBalance      float64
		wantDestinationBalance float64
		wantErr                error
	}{
		{
			name: "zero amount",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              0.00,
			},
			wantErr: service.ErrInvalidTransactionAmount,
		},
		{
			name: "negative amount",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              -0.01,
			},
			wantErr: service.ErrInvalidTransactionAmount,
		},
		{
			name: "origin wallet not found",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      uuid.UUID{},
				DestinationWalletID: walletB,
				Amount:              5.00,
			},
			wantErr: service.ErrWalletNotFound,
		},
		{
			name: "destination wallet not found",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      walletA,
				DestinationWalletID: uuid.UUID{},
				Amount:              5.00,
			},
			wantErr: service.ErrWalletNotFound,
		},
		{
			name: "access denied",
			req: service.TransferRequest{
				Issuer:              service.User{},
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              12.75,
			},
			checkBalance:           true,
			wantOriginBalance:      12.75,
			wantDestinationBalance: 45.00,
			wantErr:                service.ErrWalletAccessDenied,
		},
		{
			name: "insufficient funds",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              12.76,
			},
			checkBalance:           true,
			wantOriginBalance:      12.75,
			wantDestinationBalance: 45.00,
			wantErr:                service.ErrInsufficientFunds,
		},
		{
			name: "transaction ok",
			req: service.TransferRequest{
				Issuer:              userA,
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              12.75,
				Message:             ptrToStr("test"),
			},
			checkBalance: true,
			wantTransfer: service.Transfer{
				IssuerID:            userA.ID,
				OriginWalletID:      walletA,
				DestinationWalletID: walletB,
				Amount:              12.75,
				Message:             ptrToStr("test"),
			},
			wantOriginBalance:      0.00,
			wantDestinationBalance: 65.00,
			wantErr:                nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			setupFixtures(ctx, t, db)
			r := TransferRepository{
				DB:              db,
				WalletRepo:      &WalletRepository{DB: db},
				TransactionRepo: &TransactionRepository{DB: db},
			}

			got, err := r.CreateTransfer(ctx, tc.req)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			tc.wantTransfer.ID = got.ID
			tc.wantTransfer.Date = got.Date
			if !reflect.DeepEqual(tc.wantTransfer, got) {
				t.Fatalf("unexpected result: got: %+v, want %+v", got, tc.wantTransfer)
			}
			if !tc.checkBalance {
				return
			}
			w, err := r.WalletRepo.GetByID(ctx, tc.req.OriginWalletID)
			if err != nil {
				t.Fatalf("cannot load origin wallet: %v", err)
			}
			if got := w.Balance; got != tc.wantOriginBalance {
				t.Fatalf("unexpected origin balance: got: %v, want %v", got, tc.wantOriginBalance)
			}
			w, err = r.WalletRepo.GetByID(ctx, tc.req.DestinationWalletID)
			if err != nil {
				t.Fatalf("cannot load destination wallet: %v", err)
			}
			if got := w.Balance; got != tc.wantDestinationBalance {
				t.Fatalf("unexpected destination balance: got: %v, want %v", got, tc.wantDestinationBalance)
			}
		})
	}
}

// setupTestDB spins up a new instance of the database,
// loads the most recent schema and insert the fixtures
// Returns an open connection to the DB.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	var (
		user    = "test"
		pass    = "test"
		dbName  = "test"
		timeout = time.Second * 10
	)

	host, port := createDBContainer(t, user, pass, timeout)
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, dbName,
	)
	db := connectDB(t, dsn, timeout)
	migrateDB(t, db)

	return db
}

// createDBContainer creates a new DB docker container.
// Returns the host address and port.
func createDBContainer(t *testing.T, user, pass string, timeout time.Duration) (string, string) {
	t.Helper()
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatalf("cannot get a docker client instance: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: dbImage,
			Env:   []string{"POSTGRES_USER=" + user, "POSTGRES_PASSWORD=" + pass},
			ExposedPorts: nat.PortSet{
				"5432/tcp": struct{}{},
			},
		},
		&container.HostConfig{
			PublishAllPorts: true,
		},
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("cannot create db container: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
			t.Fatalf("cannot remove db container: %v", err)
		}
	})

	if err = cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		t.Fatalf("cannot start db container: %v", err)
	}

	l, err := cli.ContainerLogs(
		ctx,
		c.ID,
		types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		},
	)
	go func() { _, _ = io.Copy(os.Stdout, l) }()

	i, err := cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		t.Fatalf("cannot retrieve db container information: %v", err)
	}
	pm := i.NetworkSettings.Ports["5432/tcp"]

	return pm[0].HostIP, pm[0].HostPort
}

// connectDB establishes a connection to the test database.
func connectDB(t *testing.T, dsn string, timeout time.Duration) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("cannot open connection to db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tick := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout connecting to db: %v", err)
		case <-tick.C:
			if err = db.PingContext(ctx); err == nil {
				return db
			}
		}
	}
}

// migrateDB run migrations up in the specified directory.
func migrateDB(t *testing.T, db *sql.DB) {
	t.Helper()

	d, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatalf("cannot instantiate postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "", d)
	if err != nil {
		t.Fatalf("cannot instantiate migrator for schema: %v", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrations up failed: %v", err)
	}
}

func setupFixtures(ctx context.Context, t *testing.T, db *sql.DB) {
	query, err := ioutil.ReadFile(fixturesFile)
	if err != nil {
		t.Fatalf("cannot read fixtures file: %v", err)
	}

	if _, err = db.ExecContext(ctx, string(query)); err != nil {
		t.Fatalf("cannot execute fixtures: %v", err)
	}
}

func rfc3339MustParse(t *testing.T, datetime string) time.Time {
	x, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		t.Fatalf("cannot parse datetime as RFC3339: %v", err)
	}

	return x.UTC()
}

func ptrToUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

func ptrToStr(str string) *string {
	return &str
}
