package http

import "net/http"

// ListenAndServe is the same as net/http.ListenAndServe
//
// While we could let the Server type in this package handle
// starting itself up, I prefer to make that functionality
// available to my main and other packages, so I'll
// occasionally expose functions like this through package
// level variables.
// It also gives me a chance to customize what this function
// does if I ever have a need. Eg I could limit the ports
// or addresses that are valid with a custom ListenAndServe
// function if I needed to.
var ListenAndServe = http.ListenAndServe
