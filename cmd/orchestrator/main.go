package main

import (
	"arithmetic-expression-calculator/internal/config"
	"arithmetic-expression-calculator/internal/handle"
	"arithmetic-expression-calculator/internal/logger"
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

	mux := http.NewServeMux()
	repo := handle.NewRepo()

	mux.HandleFunc("/calculate", repo.AddExpressionHandleFunc)
	mux.HandleFunc("/expressions", repo.GetExpressionsHandleFunc)
	mux.HandleFunc("/expressions/", repo.GetExpressionsHandleFunc)
	mux.HandleFunc("/internal/task", repo.TaskHandleFunc)

	server := &http.Server{
		Handler:      mux,
		Addr:         ":" + strconv.Itoa(conf.ServerPort),
		WriteTimeout: defaultHTTPServerWriteTimeout,
		ReadTimeout:  defaultHTTPServerReadTimeout,
	}

	logger.Info("starting http server...")
	server.ListenAndServe()
}
