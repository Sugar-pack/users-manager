package app

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/internal/db"
	"github.com/Sugar-pack/users-manager/internal/grpcapi"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

const LogKeyTXID = "tx_id"

type App struct {
	grpcServer       *grpc.Server
	monitorTxService *MonitorTxService
}

func CreateApp(logger logging.Logger, dbConn *sqlx.DB, monitoringConf *config.Monitoring) *App {
	newTxIDCh := make(chan string)
	cancelTxIDCh := make(chan string)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
			otelgrpc.UnaryServerInterceptor(),
		),
	)
	server, err := grpcapi.CreateServer(grpcServer, dbConn, newTxIDCh, cancelTxIDCh)
	if err != nil {
		logger.WithError(err).Error("create grpc server failed")
		return nil
	}

	monitorTxService := CreateMonitorTxService(dbConn, newTxIDCh, cancelTxIDCh, monitoringConf.RollbackTimeout)

	return &App{
		grpcServer:       server,
		monitorTxService: monitorTxService,
	}
}

func (app *App) Start(logger logging.Logger, apiConf *config.API) {
	startCtx := context.Background()

	tracingProvider, err := initJaegerTracing(logger)
	if err != nil {
		logger.WithError(err).Error("init jaeger tracing failed")
		return
	}
	defer func() {
		if stopErr := tracingProvider.Shutdown(startCtx); stopErr != nil {
			logger.WithError(stopErr).Error("shutting down tracer provider failed")
		}
	}()

	grpcAddr := apiConf.Bind
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.WithError(err).WithField("addr", grpcAddr).Error("listen failed")
		return
	}

	ctx, cancelFn := context.WithCancel(startCtx)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		logger.WithField("bind_addr", grpcAddr).Trace("starting grpc server")
		if serveErr := app.grpcServer.Serve(lis); serveErr != nil {
			logger.WithError(err).Error("serve grpc server failed")
		}
		cancelFn()
		wg.Done()
	}()

	monitorCtx := logging.WithContext(ctx, logger)
	wg.Add(1)
	go func() {
		logger.Trace("starting monitor service")
		app.monitorTxService.Serve(monitorCtx)
		wg.Done()
	}()

	wg.Wait()
}

type MonitorTxService struct {
	newTxIDCh    chan string
	cancelTxIDCh chan string
	dbConn       *sqlx.DB
	timeout      time.Duration
}

func CreateMonitorTxService(dbConn *sqlx.DB,
	newTxIDCh, cancelTxIDCh chan string,
	timeout time.Duration,
) *MonitorTxService {
	return &MonitorTxService{
		newTxIDCh:    newTxIDCh,
		cancelTxIDCh: cancelTxIDCh,
		dbConn:       dbConn,
		timeout:      timeout,
	}
}

func (mtx *MonitorTxService) Serve(ctx context.Context) {
	logger := logging.FromContext(ctx)
	cancelation := make(map[string]*time.Timer)
	timeout := mtx.timeout
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case txID := <-mtx.newTxIDCh:
			timer := time.AfterFunc(timeout, func() {
				rollbackCtx, span := otel.Tracer("monitor_tx").Start(ctx, "timeout rollback")
				defer span.End()
				err := db.RollbackPrepared(rollbackCtx, mtx.dbConn, txID)
				if err != nil {
					logger.WithError(err).WithField(LogKeyTXID, txID).Error("rollback timeouted prepared tx failed")
				}
				logger.WithField(LogKeyTXID, txID).Trace("tx rollbacked due timeout")
			})
			logger.WithFields(logging.Fields{LogKeyTXID: txID, "timeout": timeout}).Trace("watching tx")
			cancelation[txID] = timer // this is not thread safe
		case txID := <-mtx.cancelTxIDCh:
			if timer, ok := cancelation[txID]; ok { // this is not thread safe
				logger.WithField(LogKeyTXID, txID).Trace("unwatching tx")
				timer.Stop()
				delete(cancelation, txID)
			}
		}
	}
}
