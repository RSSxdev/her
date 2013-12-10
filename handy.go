package handy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
)

const VERSION = "0.0.1 beta"

var (
	Config         *config
	driverName     string
	dataSourceName string
)

type Application struct {
	Route *Router
}

func (app *Application) New(config map[string]interface{}) *Application {
	Config = LoadConfig(config)
	application := &Application{Route: newRouter()}
	return application
}

func (app *Application) Connection(dsn, conn string) {
	driverName = dsn
	dataSourceName = conn
}

func (app *Application) FuncMap(tmplFunc map[string]interface{}) {
	if len(tmplFunc) > 0 {
		for k, v := range tmplFunc {
			funcMap[k] = v
		}
	}
}

func (app *Application) Start() {
	address := Config.Get("Address").String()
	if address == "" {
		address = "0.0.0.0"
	}
	port := Config.Get("Port").String()
	if port == "" {
		port = "8080"
	}
	debug := Config.Get("Debug").Bool()
	tmplPath := Config.Get("TemplatePath").String()
	listen := fmt.Sprintf("%s:%s", address, port)

	loadTemplate()
	watcher := NewWatcher()
	watcher.Listen(tmplPath)
	watcher.Notify()

	mux := http.NewServeMux()
	if debug {
		mux.Handle("/debug/pprof", http.HandlerFunc(pprof.Index))
		mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		mux.Handle("/debug/pprof/block", pprof.Handler("block"))
		mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	}
	mux.Handle("/", app.Route)

	l, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening on " + listen + "...")
	log.Fatal(http.Serve(l, mux))
}
