package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"safedeal-backend-trainee/cmd/api/handler"
	"safedeal-backend-trainee/internal/postgres"
	"safedeal-backend-trainee/pkg/log/logger"
	"syscall"
	"time"
)

func main() {
	logger := initLogger()

	st, closers := initStorages(logger)

	defer handleClosers(logger, closers)

	h := handler.New(st.p, st.o, logger)
	srv := initServer(h, "", "5000")

	const Duration = 5
	go gracefulShutdown(srv, Duration*time.Second, logger)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func handleCloser(l logger.Logger, resource string, closer io.Closer) {
	if err := closer.Close(); err != nil {
		l.Errorf("Can't close %q: %s", resource, err)
	}
}

func handleClosers(l logger.Logger, m map[string]io.Closer) {
	for n, c := range m {
		if err := c.Close(); err != nil {
			l.Errorf("Can't close %q: %s", n, err)
		}
	}
}

type storages struct {
	p *postgres.ProductStorage
	o *postgres.OrderStorage
}

func initStorages(logger logger.Logger) (*storages, map[string]io.Closer) {
	closers := make(map[string]io.Closer)

	_ = os.Chdir("../..")
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatalf("can't get path: %v", err)
	}

	db, err := postgres.New(logger, fmt.Sprintf("%s/configuration.json", pwd))
	if err != nil {
		logger.Fatalf("can't create database instance %v", err)
	}

	closers["db"] = db

	err = db.CheckConnection()
	if err != nil {
		logger.Fatalf("can't connect to database %v", err)
	}

	productStorage, err := postgres.NewProductStorage(db)
	if err != nil {
		logger.Fatalf("can't create product storage: %s", err)
	}

	closers["product_storage"] = productStorage

	orderStorage, err := postgres.NewOrderStorage(db)
	if err != nil {
		logger.Fatalf("can't create order storage: %s", err)
	}

	closers["order_storage"] = productStorage

	return &storages{productStorage, orderStorage}, closers
}

func initServer(h *handler.Handler, host string, port string) *http.Server {
	r := routes(h)
	addr := net.JoinHostPort(host, port)
	srv := &http.Server{Addr: addr, Handler: r}

	return srv
}

func routes(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	const Duration = 60

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(Duration * time.Second))

	r.Mount("/", h.Routes())

	return r
}

func gracefulShutdown(srv *http.Server, timeout time.Duration, logger logger.Logger) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Infof("Shutting down server with %s timeout", timeout)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("could not shutdown server:%v", err)
	}
}

func initLogger() logger.Logger {
	config := logger.Configuration{
		EnableConsole:     true,
		ConsoleLevel:      logger.Debug,
		ConsoleJSONFormat: true,
		EnableFile:        true,
		FileLevel:         logger.Info,
		FileJSONFormat:    true,
		FileLocation:      "log.log",
	}

	logger, err := logger.New(config, logger.InstanceZapLogger)
	if err != nil {
		log.Fatal("could not instantiate logger: ", err)
	}

	return logger
}
