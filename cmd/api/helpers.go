package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arynkh/greenlight/internal/validator"
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) //1MB limit

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() //catch unwanted fields in the JSON request body

	//decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		//use the errors.As() function to check whether the error has the type *json.SyntaxError
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		//check for syntax errors
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly-formed JSON")
		//check when the JSON value is the wrong type for the target destination
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		//if the JSON contains a field which cannot be mapped to the target destination
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		//returns if the request body exceeded our size limit of 1MB
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		//returned if we pass something that is not a non-nil pointer as the target destination
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString() returns a string value from the query string, or the provided default value if no matching key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	//Extract the value of the given key from the query string. If no key exists this will return the empty string ""
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV() reads a string value from the query string, splits it on commas and returns a slice of strings.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	//Extract the value from the query string.
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	//Use the strings.Split() function to split the string into a slice based on the , delimiter
	return strings.Split(csv, ",")
}

// The readInt() helper reads a string value from the query string and converts it to an
// integer before returning.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}
