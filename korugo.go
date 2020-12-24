package korugo

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4"
)

// Config defines the App configuration
type Config struct {

	// http route prefix. Default: "v1"
	Prefix string

	// http port. Default: 8080
	Port uint16

	// http router. Default: mux.NewRouter()
	Router *mux.Router
}

type App struct {
	Config *Config
	log    *log.Logger
}

var defaultConfig = Config{
	Prefix: "v1",
	Port:   8080,
}

// Run initializes and starts the korugo application.
// Any unset config values are automatically filled.
// See Config struct for more information.
func (app *App) Run() {

	app.log = log.New(os.Stdout, "korugo", log.LstdFlags|log.Lshortfile)

	// Initialize config defaults
	cnf := app.Config
	if cnf == nil {
		cnf = &defaultConfig
	}
	if cnf.Port == 0 {
		cnf.Port = defaultConfig.Port
	}
	if cnf.Prefix == "" {
		cnf.Prefix = defaultConfig.Prefix
	}
	if cnf.Router == nil {
		cnf.Router = mux.NewRouter()
	}
	app.Config = cnf

	// Begin the GraphQL server
	r := cnf.Router.PathPrefix(fmt.Sprintf("/%s", app.Config.Prefix)).Subrouter()
	r.Handle("/gql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r = r.WithContext(ctx)
		// srv.ServeHTTP(w, r)
	}))

	app.Config = cnf

	app.log.Printf("Starting http server on %v", app.Config.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", app.Config.Port), r)
	app.log.Fatalf("%+v", err)
}
