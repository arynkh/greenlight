package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Retrieve the "id" URL param from the current request context & convert it to an integer
func (app *application) readIDParam(r *http.Request) (int64, error) {
	//when httprouter is parsing a request, any interpolated URL parameters will be
	//stored in the request context. Use the ParamsFromContext() func to retrieve a
	//slice containing these parameter names and values.
	params := httprouter.ParamsFromContext(r.Context())

	//ByName() method gets the value of the "id" parameter from the slice.
	//The value returned by ByName() is always a string. We then convert it to a base 10 int (with a bit size of 64)
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelop map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelop, headers http.Header) error {
	//returns a []byte slice containing the encoded JSON
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json") //add header
	w.WriteHeader(status)                              //write the status code
	w.Write(js)                                        //write the JSON as the HTTP response body

	return nil
}
