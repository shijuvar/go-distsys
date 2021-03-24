package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	kithttp "github.com/go-kit/kit/transport/http"
	_ "github.com/lib/pq"
	"github.com/shijuvar/go-distsys/gokitdemo/pkg/oc"

	"github.com/shijuvar/go-distsys/gokitdemo/services/order"
	"github.com/shijuvar/go-distsys/gokitdemo/services/order/cockroachdb"
	ordersvc "github.com/shijuvar/go-distsys/gokitdemo/services/order/implementation"
	"github.com/shijuvar/go-distsys/gokitdemo/services/order/middleware"
	"github.com/shijuvar/go-distsys/gokitdemo/services/order/transport"
	httptransport "github.com/shijuvar/go-distsys/gokitdemo/services/order/transport/http"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":3000", "HTTP listen address")
	)
	flag.Parse()
	// initialize our OpenCensus configuration and defer a clean-up
	defer oc.Setup("order").Close()
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger,
			"svc", "order",
			"ts", log.DefaultTimestampUTC,
			"caller", log.DefaultCaller,
		)
	}

	level.Info(logger).Log("msg", "service started")
	defer level.Info(logger).Log("msg", "service ended")

	var db *sql.DB
	{
		var err error
		// Connect to the "ordersdb" database
		db, err = sql.Open("postgres",
			"postgresql://shijuvar@localhost:26257/ordersdb?sslmode=disable")
		if err != nil {
			level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
	}

	// Create Order Service
	var svc order.Service
	{
		repository, err := cockroachdb.New(db, logger)
		if err != nil {
			level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
		svc = ordersvc.NewService(repository, logger)
		// Add service middleware here
		// Logging middleware
		svc = middleware.LoggingMiddleware(logger)(svc)

	}
	// Create Go kit endpoints for the Order Service
	// Then decorates with endpoint middlewares
	var endpoints transport.Endpoints
	{
		endpoints = transport.MakeEndpoints(svc)
		// Add endpoint level middlewares here
		// Trace server side endpoints with open census
		endpoints = transport.Endpoints{
			Create:       oc.ServerEndpoint("Create")(endpoints.Create),
			GetByID:      oc.ServerEndpoint("GetByID")(endpoints.GetByID),
			ChangeStatus: oc.ServerEndpoint("ChangeStatus")(endpoints.ChangeStatus),
		}

	}
	var h http.Handler
	{
		ocTracing := kitoc.HTTPServerTrace()
		serverOptions := []kithttp.ServerOption{ocTracing}
		h = httptransport.NewService(endpoints, serverOptions, logger)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		level.Info(logger).Log("transport", "HTTP", "addr", *httpAddr)
		server := &http.Server{
			Addr:    *httpAddr,
			Handler: h,
		}
		errs <- server.ListenAndServe()
	}()
	level.Error(logger).Log("exit", <-errs)
	/*
			With package "github.com/oklog/run"

			var g run.Group
			{
			   server := &http.Server{
							Addr:    *httpAddr,
							Handler: h,
						}
			       g.Add(func() error {
					   return server.ListenAndServe()
					}, func(error) {
							server.Close()
						})
		    }

					{
							// set-up our signal handler
							var (
								cancelInterrupt = make(chan struct{})
								c               = make(chan os.Signal, 2)
							)
							defer close(c)

							g.Add(func() error {
								signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
								select {
								case sig := <-c:
									return fmt.Errorf("received signal %s", sig)
								case <-cancelInterrupt:
									return nil
								}
							}, func(error) {
								close(cancelInterrupt)
							})
						}

						// spawn our goroutines and wait for shutdown
						level.Error(logger).Log("exit", g.Run())
	*/
}
