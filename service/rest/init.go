package rest

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/HPInc/krypton-fs/service/config"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	fsLogger             *zap.Logger
	debugLogRestRequests bool
	// server settings
	authConfig *config.Auth

	// file name validation for s3 files. we are doing a restricted character set
	// than is originally supported by s3
	fileNameRegex = regexp.MustCompile(`^([a-z]|[A-Z]|[0-9]|[\._-])+$`)
)

const (
	// HTTP server timeouts for the REST endpoint.
	readTimeout  = (time.Second * 5)
	writeTimeout = (time.Second * 5)
)

// Represents the FS REST service.
type fsRestService struct {
	// Signal handling to support SIGTERM and SIGINT for the service.
	errChannel  chan error
	stopChannel chan os.Signal

	// Prometheus metrics reporting.
	metricRegistry *prometheus.Registry

	// Request router
	router *mux.Router

	// HTTP port on which the REST server is available.
	port int
}

// Creates a new instance of the FS REST service and initalizes the request
// router for the FS REST endpoint.
func newFsRestService() *fsRestService {
	s := &fsRestService{}

	// Initial signal handling.
	s.errChannel = make(chan error)
	s.stopChannel = make(chan os.Signal, 1)
	signal.Notify(s.stopChannel, syscall.SIGINT, syscall.SIGTERM)

	// Initialize the prometheus metric reporting registry.
	s.metricRegistry = prometheus.NewRegistry()

	s.router = initRequestRouter()
	return s
}

// Starts the HTTP REST server for the FS service and starts serving requests
// at the REST endpoint.
func (s *fsRestService) startServing() {
	// Start the HTTP REST server. http.ListenAndServe() always returns
	// a non-nil error
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        s.router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	fsLogger.Error("Received a fatal error from http.ListenAndServe",
		zap.Error(err),
	)

	// Signal the error channel so we can shutdown the service.
	s.errChannel <- err
}

// Waits for the FS REST server to be terminated - either in response to a
// system event received on the stop channel or a fatal error signal received
// on the error channel.
func (s *fsRestService) awaitTermination() {
	select {
	case err := <-s.errChannel:
		fsLogger.Error("Shutting down due to a fatal error.",
			zap.Error(err),
		)
	case sig := <-s.stopChannel:
		fsLogger.Info("Received an OS signal to shut down!",
			zap.String("Signal received: ", sig.String()),
		)
	}
}

// Init initializes the FS REST server and starts serving REST requests at the
// FS's REST endpoint.
func Init(logger *zap.Logger, settings *config.Config) {
	fsLogger = logger
	debugLogRestRequests = settings.Server.DebugRestRequests
	authConfig = &settings.Server.Auth

	s := newFsRestService()
	s.port = settings.Server.Port

	// Initialize the REST server and listen for REST requests on a separate
	// goroutine. Report fatal errors via the error channel.
	go s.startServing()
	fsLogger.Info("Started the FS REST service!",
		zap.Int("Port: ", s.port),
	)

	// Wait for the REST server to be terminated either in response to a system
	// event (like service shutdown) or a fatal error.
	s.awaitTermination()
}
