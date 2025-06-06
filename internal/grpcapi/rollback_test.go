package grpcapi

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/internal/db"
	"github.com/Sugar-pack/users-manager/internal/migrations"
	"github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	"github.com/Sugar-pack/users-manager/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type RollbackTxSuite struct {
	suite.Suite
	ctx        context.Context
	dbConn     *sqlx.DB
	dockerPool *dockertest.Pool
	pgResource *dockertest.Resource
}

func TestRollbackTxSuite(t *testing.T) {
	ts := new(RollbackTxSuite)
	suite.Run(t, ts)
}

func (ts *RollbackTxSuite) SetupSuite() {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	pool, err := dockertest.NewPool("")
	if err != nil {
		ts.T().Skipf("docker not available: %v", err)
	}
	if err = pool.Client.Ping(); err != nil {
		ts.T().Skipf("docker not available: %v", err)
	}

	dbUser := "user_db"
	dbName := "users_db"
	sslMode := "disable"
	// pulls an image, creates a container based on it and runs it
	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15.13",
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
		log.Fatalf("Could not start pgResource: %s", err)
	}

	hostAndPort := pgResource.GetHostPort("5432/tcp")
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

	err = migrations.Apply(ctx, dbConf)
	if err != nil {
		log.Fatalf("apply migrations failed: '%s'", err)
	}

	ts.ctx = ctx
	ts.pgResource = pgResource
	ts.dockerPool = pool
	ts.dbConn = dbConn
}

func (ts *RollbackTxSuite) TearDownSuite() {
	ctx := ts.ctx
	dbConn := ts.dbConn
	if disconnectErr := db.Disconnect(ctx, dbConn); disconnectErr != nil {
		log.Fatalf("disconnect failed: '%s'", disconnectErr)
	}

	dockerPool := ts.dockerPool
	pgResource := ts.pgResource
	if purgeErr := dockerPool.Purge(pgResource); purgeErr != nil {
		log.Fatalf("Could not purge pgResource: %s", purgeErr)
	}
}

func (ts *RollbackTxSuite) TestRollback_OK() {
	t := ts.T()
	ctx := ts.ctx
	logger := logging.FromContext(ctx)
	dbConn := ts.dbConn

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

	grpcConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err == nil {
		grpcConn.Connect()
	}
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		assert.NoError(t, grpcConn.Close())
	})

	testTx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	userName := "foobar"
	userID := uuid.New()
	testUser := &db.User{
		CreatedAt: now,
		Name:      userName,
		ID:        userID,
	}
	if err = db.CreateUser(ctx, testTx, testUser); err != nil {
		t.Fatal(err)
	}
	txID := uuid.New()
	if err = db.PrepareTransaction(ctx, testTx, txID.String()); err != nil {
		t.Fatal(err)
	}

	// it really doesn't matter what we do with PREPARED TRANSACTION
	// neither COMMIT nor ROLLBACK has effect
	// here, we just release testTx
	if err = testTx.Rollback(); err != nil {
		t.Fatal(err)
	}

	txClient := distributedtx.NewDistributedTxServiceClient(grpcConn)
	resp, err := txClient.Rollback(ctx, &distributedtx.TxToRollback{TxId: txID.String()})
	assert.NoError(t, err, "unexpected error")
	assert.NotNil(t, resp, "unexpected grpc response")

	var cancelTxID string
	select {
	case cancelTxID = <-cancelTxIDCh:
	default:
		break
	}
	assert.NotEmpty(t, cancelTxID, "unexpected cancel tx id")

	_, err = fetchPreparedTx(t, ctx, dbConn, cancelTxID)
	assert.ErrorIs(t, err, sql.ErrNoRows, "unexpected error, expected NoRows")

	dbUser, err := fetchDBUser(t, dbConn, userID)
	assert.ErrorIs(t, err, sql.ErrNoRows, "unexpected error, expected NoRows")
	assert.Empty(t, dbUser, "dbUser must be empty")
}
