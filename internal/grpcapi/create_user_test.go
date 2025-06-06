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
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const buffSize = 1024 * 1024

type CreateUserSuite struct {
	suite.Suite
	ctx        context.Context
	dbConn     *sqlx.DB
	dockerPool *dockertest.Pool
	pgResource *dockertest.Resource
}

func TestCreateUserSuite(t *testing.T) {
	ts := new(CreateUserSuite)
	suite.Run(t, ts)
}

func (ts *CreateUserSuite) SetupSuite() {
	ctx := context.Background()
	logger := logging.GetLogger()
	ctx = logging.WithContext(ctx, logger)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dbUser := "user_db"
	dbName := "users_db"
	sslMode := "disable"
	// pulls an image, creates a container based on it and runs it
	pgResource, err := pool.RunWithOptions(&dockertest.RunOptions{
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

func (ts *CreateUserSuite) TearDownSuite() {
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

func (ts *CreateUserSuite) TestCreateUser_OK() {
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

	grpcConn, err := grpc.NewClient("",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		assert.NoError(t, grpcConn.Close())
	})

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

	dbTXID, err := fetchPreparedTx(t, ctx, dbConn, monitorTXID)
	if err != nil {
		t.Fatalf("select prepared tx failed: '%s'", err)
	}
	assert.NotEmpty(t, dbTXID, "unexpected db tx id value")
}

func fetchPreparedTx(t *testing.T, ctx context.Context, dbConn *sqlx.DB, txID string) (string, error) {
	t.Helper()
	var dbTxId string
	err := dbConn.QueryRowxContext(ctx, `SELECT gid FROM pg_prepared_xacts WHERE gid = $1`, txID).Scan(&dbTxId)
	return dbTxId, err
}
