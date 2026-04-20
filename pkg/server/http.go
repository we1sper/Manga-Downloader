package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/we1sper/Manga-Downloader/pkg/log"
)

type HttpServer struct {
	address string
	port    int

	server   http.Server
	serveMux http.ServeMux

	sigChan         chan os.Signal
	preHandleHooks  []func(w http.ResponseWriter, r *http.Request)
	onShutdownHooks []func()
}

func NewHttpServer(address string, port int) *HttpServer {
	return &HttpServer{
		address:         address,
		port:            port,
		server:          http.Server{},
		serveMux:        http.ServeMux{},
		sigChan:         make(chan os.Signal, 1),
		preHandleHooks:  make([]func(w http.ResponseWriter, r *http.Request), 0),
		onShutdownHooks: make([]func(), 0),
	}
}

func (srv *HttpServer) RegisterHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) *HttpServer {
	srv.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		// Step 1: execute pre-handle hooks.
		for _, hook := range srv.preHandleHooks {
			hook(w, r)
		}
		// Step 2: execute handler.
		handler(w, r)
	})
	srv.server.Handler = &srv.serveMux
	log.Infof("[server][http] pattern %s registered", pattern)
	return srv
}

func (srv *HttpServer) RegisterPreHandleHook(hook func(http.ResponseWriter, *http.Request)) *HttpServer {
	srv.preHandleHooks = append(srv.preHandleHooks, hook)
	return srv
}

func (srv *HttpServer) RegisterOnShutdownHook(hook func()) *HttpServer {
	srv.onShutdownHooks = append(srv.onShutdownHooks, hook)
	return srv
}

func (srv *HttpServer) Start() {
	srv.listenSignals()

	srv.server.Addr = fmt.Sprintf("%s:%d", srv.address, srv.port)
	log.Infof("[server][http] listening on %s", srv.server.Addr)
	if err := srv.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("[server][http] closed")
		} else {
			log.Warnf("[server][http] terminated: %v", err)
		}
	}
}

func (srv *HttpServer) Stop() {
	if err := srv.server.Shutdown(context.Background()); err != nil {
		log.Errorf("[server][http] error occurred while shutting down the server: %v", err)
	}
}

func (srv *HttpServer) AllowCORS() *HttpServer {
	return srv.RegisterPreHandleHook(srv.setCORSHeaders)
}

func (srv *HttpServer) listenSignals() {
	signal.Notify(srv.sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-srv.sigChan
		for _, hook := range srv.onShutdownHooks {
			hook()
		}
		srv.Stop()
	}()
}

func (srv *HttpServer) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}
