package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	oczipkin "contrib.go.opencensus.io/exporter/zipkin"
	"github.com/go-kit/kit/log"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	kittransport "github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/oklog/run"
	zipkin "github.com/openzipkin/zipkin-go"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"

	"github.com/shijuvar/go-distsys/examples/gokit/endpoints"
	svchttp "github.com/shijuvar/go-distsys/examples/gokit/http"
	"github.com/shijuvar/go-distsys/examples/gokit/service"
)

const (
	serviceName = "oc-gokit-example"
	zipkinURL   = "http://localhost:9411/api/v2/spans"
)

func main() {
	// Set-up our contextual logger.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger, "svc", serviceName)
	}

	// Set-up our OpenCensus instrumentation with Zipkin backend
	{
		var (
			reporter         = httpreporter.NewReporter(zipkinURL)
			localEndpoint, _ = zipkin.NewEndpoint(serviceName, ":0")
			exporter         = oczipkin.NewExporter(reporter, localEndpoint)
		)
		defer reporter.Close()

		// Always sample our traces for this demo.
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

		// Register our trace exporter.
		trace.RegisterExporter(exporter)
	}

	// Set-up our service.
	var handler http.Handler
	{
		// Create our User service implementation.
		svc := service.Service{}

		// Create our Go kit Endpoints.
		endpoints := endpoints.Endpoints{
			PostUser: endpoints.MakePostUserEndpoint(svc),
		}

		// Wrap our service endpoints with OpenCensus tracing middleware.
		endpoints.PostUser = kitoc.TraceEndpoint("gokit:endpoint postuser")(endpoints.PostUser)

		// Set-up our Go kit HTTP transport options.
		var serverOptions []httptransport.ServerOption
		serverOptions = append(serverOptions, httptransport.ServerErrorHandler(kittransport.NewLogErrorHandler(logger)))
		serverOptions = append(serverOptions, kitoc.HTTPServerTrace())

		// Create our HTTP transport handler.
		handler = svchttp.NewHTTPHandler(endpoints, serverOptions...)
	}

	// run.Group manages our goroutine lifecycles
	// see: https://www.youtube.com/watch?v=LHe1Cb_Ud_M&t=15m45s
	var g run.Group
	{
		// Set-up our HTTP service.
		var (
			listener, _ = net.Listen("tcp", ":0") // dynamic port assignment
			addr        = listener.Addr().String()
		)
		g.Add(func() error {
			logger.Log("msg", "service start", "transport", "http", "address", addr)
			return http.Serve(listener, handler)
		}, func(error) {
			listener.Close()
		})
	}
	{
		// Set-up our signal handler.
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

	// Spawn our Go routines and wait for shutdown.
	logger.Log("exit", g.Run())
}
