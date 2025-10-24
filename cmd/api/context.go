package main

import (
	"context"
	"net/http"

	"github.com/arynkh/greenlight/internal/data"
)

// Define a custom context key type
type contextKey string

// Convert the string "user" to a contextKey type and assign it to teh userContextKey constant. This will be used as the key when storing and retrieving the authenticated user information from the request context
const userContextKey = contextKey("user")

// The contextSetUser() method returns a new copy of the request with the provided
// User struct added to the context.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetUser() retrieves the user information from the request context. This will only be used when we logically expect there to be a User struct value in the context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
