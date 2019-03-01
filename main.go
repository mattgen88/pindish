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

	viper.SetEnvPrefix("pindish")

	// Pinterest data
	viper.BindEnv("app_id")
	viper.BindEnv("app_secret")

	// Api environment
	viper.BindEnv("port")
	viper.BindEnv("host")

	// Database environment
	viper.BindEnv("dbport")
	viper.BindEnv("dbhost")

	// Set up logging
	log.SetFormatter(&log.TextFormatter{})

	// Set up routes
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)

	// Middleware
	loggedRouter := gorilla.LoggingHandler(os.Stdout, r)

	// Start
	h := net.JoinHostPort(viper.GetString("host"), viper.GetString("port"))
	http.ListenAndServe(h, loggedRouter)
}
