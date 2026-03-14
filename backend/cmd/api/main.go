package main

import (
	"log"
	"main/internal/bootstrap"
	"main/internal/config"
	"main/internal/db"
	"main/internal/redis"
	"main/internal/server"
	"main/internal/smtp"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	dbConn := db.NewDbConnection(cfg.DB.ConnString())
	defer dbConn.Close()

	db.MigrateDB(dbConn)

	redisConn := redis.NewRedisConnection(cfg.Redis)
	defer redisConn.Close()

	smtpClient := smtp.NewSMTPclient(cfg.SMTP)

	adapters := bootstrap.NewAdapters(dbConn, redisConn, smtpClient)
	services := bootstrap.NewServices(adapters)
	handlers := bootstrap.NewHandlers(services)
	middlewares := bootstrap.NewMiddlewares(services, &cfg)

	srv := server.New(cfg.AppPort)
	srv.RegisterHandlersAndMiddlewares(handlers, middlewares)
	srv.RegisterStaticRoute()
	srv.RegisterSpaRoute("./public")
	srv.RegisterSwaggerRoute()

	log.Printf("Server running on :%s", cfg.AppPort)
	log.Fatal(srv.Start(true))
}
