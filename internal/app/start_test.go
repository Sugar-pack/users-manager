package app

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func TestAppStart_StopImmediately(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	defer func() { _ = dbConn.Close() }()
	monitorConf := &config.Monitoring{RollbackTimeout: 10 * time.Millisecond}
	app := CreateApp(logging.GetLogger(), dbConn, monitorConf)
	apiConf := &config.API{Bind: "127.0.0.1:0"}
	done := make(chan struct{})
	go func() {
		app.Start(logging.GetLogger(), apiConf)
		close(done)
	}()
	time.Sleep(10 * time.Millisecond)
	app.grpcServer.Stop()
	<-done
}
