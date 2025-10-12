package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/arynkh/greenlight/internal/data"
	"github.com/arynkh/greenlight/internal/mailer"

	_ "github.com/lib/pq"
)

const version = "1.0.0" //app version number

// holds all config settings for the app.
type config struct {
	port int
	env  string //(dev, staging, prod, etc)
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// holds dependencies for our HTTP handlers, helpers & middleware
type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer *mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	//use the value of the GREENLIGHT_DB_DSN environment variable as the default value for the db-dsn command-line flag
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// Create command-line flags to read the settings values into the config struct.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	//Read the SMTP server configuration settings into the config struct, using the Mailtrap settings as the default values.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "d546433426c99e", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "395abe4d24d984", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <noreply@greenlight.arynhead.net>", "SMTP sender")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	//call the openDB() helper function to create the connection pool, passing in the config struct as an argument.
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	//defer a call to db.Close() so that the connection pool is closed before the main() function exits.
	defer db.Close()

	logger.Info("database connection pool established")

	//Initialize a new mailer.Mailer instance using the settings from the command line flags
	mailer, err := mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer,
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// the openDB() function returns a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
	//use sql.Open() to create an empty connection pool, using the DSN from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	//set the maximum number of open (in-use + idle) connections in the pool
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// set the maximum number of idle connections in the pool
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// set the maximum idle timeout for connections in the pool.
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	//create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//use PingContext() to establish a new connection to the database, passing in the context we created above as a param.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	//return the sql.DB connection pool
	return db, nil
}
