package main

import (
	"net"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/mattgen88/pindish/handlers"

	"database/sql"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

func main() {
	// Setup environment
	viper.AutomaticEnv()

	// Api environment from heroku
	viper.BindEnv("port")
	viper.BindEnv("host")
	viper.BindEnv("database_url")

	// Api environment from env file
	viper.SetEnvPrefix("pindish")

	// Pinterest data
	viper.BindEnv("app_id")
	viper.BindEnv("app_secret")
	viper.BindEnv("frontend_uri")
	viper.BindEnv("mock")

	// Set up logging
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)

	// Connect to postgres
	db, err := sql.Open("postgres", viper.GetString("database_url"))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	h := handlers.Handlers{
		DB: db,
	}

	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/", h.HomeHandler)
	r.HandleFunc("/auth", h.AuthHandler)
	r.HandleFunc("/catch", h.CatchHandler)
	r.HandleFunc("/boards", handlers.AuthRequired(h.BoardsHandler))

	headersOk := gorilla.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	originsOk := gorilla.AllowedOrigins([]string{viper.GetString("frontend_url")})
	methodsOk := gorilla.AllowedMethods([]string{"GET", "HEAD", "OPTIONS"})

	corsRouter := gorilla.CORS(headersOk, originsOk, methodsOk)(r)

	// Middleware
	loggedRouter := gorilla.LoggingHandler(os.Stdout, corsRouter)

	log.Infof("Starting on host %s port %s", viper.GetString("host"), viper.GetString("port"))

	// Start
	http.ListenAndServe(net.JoinHostPort(viper.GetString("host"), viper.GetString("port")), loggedRouter)
}
