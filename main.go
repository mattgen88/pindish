package main

import (
	"net"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/mattgen88/pindish/handlers"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Setup environment
	viper.AutomaticEnv()

	// Api environment from heroku
	viper.BindEnv("port")
	viper.BindEnv("host")

	// Api environment from env file
	viper.SetEnvPrefix("pindish")

	// Pinterest data
	viper.BindEnv("app_id")
	viper.BindEnv("app_secret")

	// Database environment
	viper.BindEnv("dbport")
	viper.BindEnv("dbhost")

	// Set up logging
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)

	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)

	// Middleware
	loggedRouter := gorilla.LoggingHandler(os.Stdout, r)

	log.Infof("Starting on host %s port %s", viper.GetString("host"), viper.GetString("port"))

	// Start
	h := net.JoinHostPort(viper.GetString("host"), viper.GetString("port"))
	http.ListenAndServe(h, loggedRouter)
}
