package gqlerrors

import (
	"fmt"
	"net/http"
)

var Internal error = fmt.Errorf("%v", http.StatusInternalServerError)
var NotFound error = fmt.Errorf("%v", http.StatusNotFound)
var Conflict error = fmt.Errorf("%v", http.StatusConflict)
var Unauthorized error = fmt.Errorf("%v", http.StatusUnauthorized)
var BadRequest error = fmt.Errorf("%v", http.StatusBadRequest)
