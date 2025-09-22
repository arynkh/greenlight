package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	//Declare an HTTP server which listens on the port provided in the config struct
	//uses the httprouter instance returned by the app.routes() as the srv handler, and writes any log messages to the
	//structures logger at Error level
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	//Start a background routine
	go func() {
		//Create a quit (buffered) channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		//Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay them to the quit channel.
		//Any other signals will not be caught by signal.Notify() and will retain their default behavior
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		//Read the signal from the quit channel. This code will block until a signal is received
		s := <-quit

		//log message to say that the signal has been caught
		app.logger.Info("caught signal", "signal", s.String())

		//Exit the application with a 0 (success) status code
		os.Exit(0)
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	return srv.ListenAndServe()
}
