package main

import (
	"arithmetic-expression-calculator/internal/config"
	"arithmetic-expression-calculator/internal/handle"
	"arithmetic-expression-calculator/internal/logger"
	"arithmetic-expression-calculator/internal/models"
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	const (
		defaultHTTPServerWriteTimeout = time.Second * 15
		defaultHTTPServerReadTimeout  = time.Second * 15
	)
	logger.Info("start main orchestrator")
	logger.Info("reading config...")
	conf, err := config.LoadFromEnv()
	if err != nil {
		logger.Error("failed to read config")
		os.Exit(1)
	}
	//---------------------- todo in repo
	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "store.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = models.CreateTables(ctx, db); err != nil {
		panic(err)
	}

	//----------------------

	mux := http.NewServeMux()
	repo := handle.NewRepo(db)

	mux.Handle("/calculate", repo.AuthMiddleware(http.HandlerFunc(repo.AddExpressionHandleFunc)))
	mux.Handle("/expressions", repo.AuthMiddleware(http.HandlerFunc(repo.GetExpressionsHandleFunc)))
	mux.Handle("/expressions/", repo.AuthMiddleware(http.HandlerFunc(repo.GetExpressionsHandleFunc)))
	mux.HandleFunc("/internal/task", repo.TaskHandleFunc)

	mux.HandleFunc("/register", repo.AddRegistration)
	mux.HandleFunc("/login", repo.Login)

	server := &http.Server{
		Handler:      mux,
		Addr:         ":" + strconv.Itoa(conf.ServerPort),
		WriteTimeout: defaultHTTPServerWriteTimeout,
		ReadTimeout:  defaultHTTPServerReadTimeout,
	}

	logger.Info("starting http server...")
	server.ListenAndServe()
}
