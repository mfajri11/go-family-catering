package app

import (
	"context"
	"family-catering/config"
	v1 "family-catering/internal/handler/http/v1"
	"family-catering/internal/repository"
	"family-catering/pkg/db/postgres"
	"family-catering/pkg/db/redis"
	"family-catering/pkg/logger"
	"family-catering/pkg/server"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
)

// @title           Family catering API
// @version         1.0
// @description     API documentation for Family-catering service.
// @termsOfService  http://swagger.io/terms/

//	@contact.name	Family catering Support
//	@contact.url	http://www.family-catering/support
//	@contact.email	support.family-catering@example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9000
//	@BasePath	/api/v1

//	@securityDefinitions.apiKey	BearerAuth
//	@in							header
//	@name						Authorization
func Run() error {
	cronTabEveryDayAt5pm00 := "0 17 * * *"

	cfg := config.Cfg()

	pg, err := postgres.New(
		cfg.Postgres.URL(),
		postgres.WithMaxIdleConns(cfg.Postgres.IdleConnection),
		postgres.WithMaxLifeTime(cfg.Postgres.ConnectionMaxLifeTime),
		postgres.WithMaxOpenConnection(cfg.Postgres.OpenConnection))
	if err != nil {
		err = fmt.Errorf("app.Run: %w", err)
		return err
	}

	redis, err := redis.New(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		redis.WithDatabaseName(cfg.Redis.DataBaseName),
		redis.WithMaxRetries(cfg.Redis.MaxRetries),
		redis.WithPoolSize(cfg.Redis.PoolSize),
	)
	if err != nil {
		logger.Fatal(err, "can't connect to redis")
	}
	// cron
	orderRepo := repository.NewOrderRepository(pg)
	jobRunner := cron.New()
	jobRunner.AddFunc(cronTabEveryDayAt5pm00, func() {
		logger.Info("cron cancelUnpaidOrder start running")
		nAffected, err := orderRepo.CancelUnpaidOrder(context.Background())
		if err != nil {
			logger.Error(err, "error execute cron cancelUnpaidOrder: %s", err.Error())
		} else {
			logger.Info("cron success execute, # affected: %d", nAffected)
		}
	})
	jobRunner.Start()

	//handler
	v1 := v1.NewRouter(pg, redis)
	srv := server.New(
		cfg.Server.Addr(),
		v1,
		server.WithReadTimeout(cfg.Server.ReadTimeout),
		server.WithWriteTimeout(cfg.Server.WriteTimeout),
		server.WithShutdownTimeout(cfg.Server.ShutDownTimeout))

	srv.Start()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Info("app.Run - os.Signal: " + s.String())
	case err = <-srv.Notify():
		logger.Error(err, "app.Run: "+err.Error())
	}

	err = srv.Shutdown()
	if err != nil {
		err = fmt.Errorf("app.Run: %w", err)
		logger.Error(err, "error shutdown server")
		return err
	}

	return nil
}
