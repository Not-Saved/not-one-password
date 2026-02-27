package main

import (
	"log"
	"main/internal/bootstrap"
	"main/internal/config"
	"main/internal/db"
	"main/internal/server"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	dbConn := db.NewDbConnection(cfg.DB.ConnString())
	defer dbConn.Close()

	repos := bootstrap.NewRepositories(dbConn)
	services := bootstrap.NewServices(repos)
	handlers := bootstrap.NewHandlers(services)
	middlewares := bootstrap.NewMiddlewares(services)

	srv := server.New(cfg.AppPort)
	srv.RegisterHandlersAndMiddlewares(handlers, middlewares)
	srv.RegisterStaticRoute()
	srv.RegisterSpaRoute("./public")
	srv.RegisterSwaggerRoute()

	log.Printf("Server running on :%s", cfg.AppPort)
	log.Fatal(srv.Start())
}
