package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//will always be run in the event of a panic
		defer func() {
			//built in reover func that checks if a panic occurred.
			pv := recover()
			if pv != nil {
				//if there is a panic, close the connection
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	//If rate limiting is not enabled, return the next handler in the chain with no further action
	if !app.config.limiter.enabled {
		return next
	}

	//Define a client struct to hold the rate limiter and last seen time for each client
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	//Declare a mutex and map to hold the client's IP addresses and rate limiters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			//Lock the mutex to prevent any rate limiter checks from happening while the cleanup is taking place
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			//Unlock the mutex when the cleanup is complete
			mu.Unlock()
		}
	}()

	//The function returns is a closure, which 'closes over' the limiter variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Use the realip.FromRequest() function to get the client's IP address
		ip := realip.FromRequest(r)

		//Lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		// Check to see if the IP address already exists in the map. If it doesn't, then
		// initialize a new client struct to the map
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
		}

		// Update the last seen time for the client
		clients[ip].lastSeen = time.Now()

		// Call the Allow() method on the rate limiter for the current IP address. If the request isn't allowed
		// unlock the mutex and send the rateLimitExceededResponse() helper to return a 429 Too Many
		// Requests response
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		//Unlock the mutex before calling the next handler in the chain
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
