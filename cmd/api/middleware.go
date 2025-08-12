package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//will always be run in the event of a panic
		defer func() {
			pv := recover()
			if pv != nil {
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
