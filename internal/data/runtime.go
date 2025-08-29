package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// define an error that our UnmarshalJSON() method can return if we're unable to parse or convert the JSON string successfully
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// implement a UnmarshalJSON() method on the Runtime type so that it satisfies the json.Unmarshaler interface.
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	//split the string to isolate the part containing the number
	parts := strings.Split(unquotedJSONValue, " ")
	//sanity check the parts of the string to make sure it was in the expected format.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	//convert the int32 to a Runtime type & assign this to the receiver. We use the * operator to dereference the receiver
	// (which is a pointer to a Runtime type) in order to set the underlying value of the pointer.
	*r = Runtime(i)

	return nil
}

// This should return the JSON-encoded value for the movie runtime
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJSONValue := strconv.Quote(jsonValue) //wrap the string in double quotes

	//convert the quoted string value to a byte slice and return it
	return []byte(quotedJSONValue), nil
}
