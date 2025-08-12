package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// This should return teh JSON-encoded value for the movie runtime
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	quotedJSONValue := strconv.Quote(jsonValue) //wrap the string in double quotes

	//convert the quoted string value to a byte slice and return it
	return []byte(quotedJSONValue), nil
}
