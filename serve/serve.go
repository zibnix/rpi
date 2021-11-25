package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	ProdPy  = "rpi.py"
	DebugPy = "debug.py"

	serverTimeout = 15 * time.Second
	dataSlug      = "data"
	pprofSlug     = "pprof"
)

var (
	// flags
	debug    bool
	interval int64
	tempPin  uint // BCM GPIO
)

func init() {
	flag.BoolVar(&debug, "debug", false, "if set, uses debug.py by default rather than trying to read from sensors")
	flag.Int64Var(&interval, "i", 5, "interval in seconds on which to read from the sensors")
	flag.UintVar(&tempPin, "tpin", 5, "BCM GPIO data pin number to read from for a DHT11 temp/humidity sensor")
	flag.Parse()
}

func main() {
	expvar.NewString("go_version").Set(runtime.Version())

	scriptFile := ProdPy
	if debug {
		scriptFile = DebugPy
	}

	ctx, cancel := context.WithCancel(context.Background())

	sensors := NewSensors(ctx, scriptFile, time.Duration(interval)*time.Second, tempPin)
	go sensors.Loop()

	server := newServer(":8080", handler(sensors))
	offline := awaitShutdown(cancel, server)

	log.Printf("Sensor data server listening on %v", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("ListenAndServe: %v", err)
	}

	<-offline
	log.Println("Sensor data server offline")

	sensors.Stop()
	log.Println("Sensor data server stoppped sensing, now exiting...")
}

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:           addr,
		Handler:        h,
		MaxHeaderBytes: 4096,
		ReadTimeout:    serverTimeout,
		WriteTimeout:   serverTimeout,
	}
}

func awaitShutdown(cancel context.CancelFunc, servers ...*http.Server) chan struct{} {
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		// interrupt signal sent from terminal
		signal.Notify(sigint, os.Interrupt)
		// sigterm signal sent from supervisor
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		log.Println("Shutdown signal received")
		cancel()

		bg := context.Background()

		var wg sync.WaitGroup
		for _, server := range servers {
			wg.Add(1)
			go func(server *http.Server) {
				defer wg.Done()
				ctx, _ := context.WithTimeout(bg, serverTimeout)
				if err := server.Shutdown(ctx); err != nil {
					log.Printf("Error while shutting down server %s: %v", server.Addr, err)
				}
			}(server)
		}

		wg.Wait()

		close(idleConnsClosed)
	}()
	return idleConnsClosed
}

func handler(s Sensors) http.Handler {
	mux := http.NewServeMux()

	dataPath := "/" + dataSlug
	mux.Handle(dataPath, MethodChecker(GET, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		JSONResp(w, s.Data())
	})))

	mux.Handle("/", http.DefaultServeMux)

	return mux
}
