package app

/*
import (
	"arithmetic-expression-calculator/internal/config"
	"arithmetic-expression-calculator/internal/storage"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type App struct {
	server *http.Server
}

func New(conf *config.Config) (*App, error) {
	const (
		defaultHTTPServerWriteTimeout = time.Second * 15
		defaultHTTPServerReadTimeout  = time.Second * 15
	)

	app := new(App)

	repo := storage.New()
	mux := http.NewServeMux()

	activeexpressionHandler := http.HandlerFunc(activeexpression.MakeHandler(activeexpression.NewSvc(repo)))
	calcexpressionHandler := http.HandlerFunc(calcexpression.MakeHandler(calcexpression.NewSvc(repo)))

	mux.Handle("/expressions", activeexpressionHandler)
	mux.Handle("/expression", calcexpressionHandler)

	app.server = &http.Server{
		Handler:      mux,
		Addr:         ":" + strconv.Itoa(conf.ServerPort),
		WriteTimeout: defaultHTTPServerWriteTimeout,
		ReadTimeout:  defaultHTTPServerReadTimeout,
	}

	return app, nil
}

func (a *App) Run() error {
	//logger.Info("starting http server...")
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server was stop with err: %w", err)
	}
	//logger.Info("server was stop")

	return nil
}

func (a *App) stop(ctx context.Context) error {
	//logger.Info("shutdowning server...")
	err := a.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server was shutdown with error: %w", err)
	}
	//logger.Info("server was shutdown")
	return nil
}

func (a *App) GracefulStop(serverCtx context.Context, sig <-chan os.Signal, serverStopCtx context.CancelFunc) {
	<-sig
	var timeOut = 30 * time.Second
	shutdownCtx, shutdownStopCtx := context.WithTimeout(serverCtx, timeOut)

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			//logger.Error("graceful shutdown timed out... forcing exit")
			os.Exit(1)
		}
	}()

	err := a.stop(shutdownCtx)
	if err != nil {
		//logger.Error("graceful shutdown timed out... forcing exit")
		os.Exit(1)
	}
	serverStopCtx()
	shutdownStopCtx()
}
*/