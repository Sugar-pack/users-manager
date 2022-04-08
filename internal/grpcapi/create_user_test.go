package grpcapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/internal/db"
	"github.com/Sugar-pack/users-manager/internal/migrations"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const buffSize = 1024 * 1024

func TestUsersService_CreateUser(t *testing.T) {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	// Prepare test environment. Look for end-section below
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dbUser := "user_db"
	dbName := "users_db"
	sslMode := "disable"
	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.2",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"POSTGRES_HOST_AUTH_METHOD=trust",
			"listen_addresses = '*'",
		},
		Cmd: []string{"postgres", "-c", "log_statement=all", "-c", "log_destination=stderr", "--max_prepared_transactions=100"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	defer func() {
		if purgeErr := pool.Purge(resource); purgeErr != nil {
			log.Fatalf("Could not purge resource: %s", purgeErr)
		}
	}()

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbHost, dbPort, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		log.Fatalf("split host-port '%s' failed: '%s'", hostAndPort, err)
	}
	dbConf := &config.DB{
		ConnString:       fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbName, sslMode),
		MigrationDirPath: "../../sql-migrations",
		MigrationTable:   "migrations",
		MaxOpenConns:     20,
		ConnMaxLifetime:  10 * time.Second,
	}

	var dbConn *sqlx.DB
	pool.MaxWait = 30 * time.Second
	if err = pool.Retry(func() error {
		dbConn, err = db.Connect(ctx, dbConf)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	defer func() {
		if disconnectErr := db.Disconnect(ctx, dbConn); disconnectErr != nil {
			log.Fatalf("disconnect failed: '%s'", err)
		}
	}()

	err = migrations.Apply(ctx, dbConf)
	if err != nil {
		log.Fatalf("apply migrations failed: '%s'", err)
	}
	// Test environment prepared

	newTxIDCh := make(chan string, 1)
	cancelTxIDCh := make(chan string, 1)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
		),
	)

	grpcapi, err := CreateServer(grpcServer, dbConn, newTxIDCh, cancelTxIDCh)

	listener := bufconn.Listen(buffSize)
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
	go func() {
		if err = grpcapi.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	grpcConn, err := grpc.DialContext(ctx, "",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcConn.Close()

	usersClient := usersPb.NewUsersClient(grpcConn)

	createdUser, err := usersClient.CreateUser(ctx, &usersPb.NewUser{Name: "e2e"})
	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, createdUser, "unexpected grpc response") // can not assert certain values, because both of them are UUID.
	// need to refactor it, and use DI for UUID generation

	var monitorTXID string
	select {
	case monitorTXID = <-newTxIDCh:
	default:
		break
	}
	assert.NotEmpty(t, monitorTXID, "unexpected monitor tx id")

	dbTXID, err := fetchPrepapedTx(ctx, dbConn, monitorTXID)
	if err != nil {
		t.Fatalf("select prepared tx failed: '%s'", err)
	}
	assert.NotEmpty(t, dbTXID, "unexpected db tx id value")
}

func fetchPrepapedTx(ctx context.Context, dbConn *sqlx.DB, txID string) (string, error) {
	var dbTxId string
	err := dbConn.QueryRowxContext(ctx, `SELECT gid FROM pg_prepared_xacts WHERE gid = $1`, txID).Scan(&dbTxId)
	return dbTxId, err
}
