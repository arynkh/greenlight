package main

import (
	"context"
	"errors"
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

	//Create a shutdownError channel. Use this to receive any errors returned by the graceful Shutdown() function
	shutdownError := make(chan error)

	//Start a background routine
	go func() {
		//Create a quit (buffered) channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		//Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals and relay them to the quit channel.
		//Any other signals will not be caught by signal.Notify() and will retain their default behavior
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		//Read the signal from the quit channel. This code will block until a signal is received
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//Call Shutdown() on our server, passing in the context.
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		// Call Wait() to block until our WaitGroup counter is zero --- essentially
		// blocking until the background goroutines have finished. Then we return nil on
		// the shutdownError channel, to indicate that the shutdown completed without
		// any issues.
		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	//Wait to receive the return value from Shutdown() on the shutdownError channel. If the return value is an error,
	//we know that there was a problem with the graceful shutdown and we return the error
	err = <-shutdownError
	if err != nil {
		return err
	}

	//By here we know that the graceful shutdown was completed successfully and we log a "stopped server" message
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
