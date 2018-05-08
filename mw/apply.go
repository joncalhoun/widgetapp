package mw

import "net/http"

// Apply will take an initial handler, then apply each middleware function to it
// so that they run in the order provided. That is, if you pass in:
//
//   Apply(h, mw1, mw2)
//
// mw1 will run before mw2, then mw2 will run before h is called.
func Apply(h http.Handler, mws ...func(http.Handler) http.HandlerFunc) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ApplyFn is a convenience wrapper for Apply that accepts an http.HandlerFunc
// instead of an http.Handler. This makes it easier to call with a function
// that isn't already explicitly typed as an http.HandlerFunc, eg:
//
//   func Demo(w http.ResponseWriter, r *http.Request) { ... }
//   ApplyFn(Demo, ...)
//
func ApplyFn(fn http.HandlerFunc, mws ...func(http.Handler) http.HandlerFunc) http.Handler {
	return Apply(fn, mws...)
}
