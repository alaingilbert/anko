//go:build !appengine

package packages

import (
	"net/http"
	"reflect"
)

func init() {
	Packages.Insert("net/http", PackageMap{
		"DefaultClient":     http.DefaultClient,
		"DefaultServeMux":   http.DefaultServeMux,
		"DefaultTransport":  http.DefaultTransport,
		"Handle":            http.Handle,
		"HandleFunc":        http.HandleFunc,
		"ListenAndServe":    http.ListenAndServe,
		"ListenAndServeTLS": http.ListenAndServeTLS,
		"NewRequest":        http.NewRequest,
		"NewServeMux":       http.NewServeMux,
		"Serve":             http.Serve,
		"SetCookie":         http.SetCookie,
		"Get":               http.Get,
		"Post":              http.Post,
		"PostForm":          http.PostForm,
	})
	PackageTypes.Insert("net/http", PackageMap{
		"Client":       http.Client{},
		"Cookie":       http.Cookie{},
		"CookieJar":    reflect.TypeOf((*http.CookieJar)(nil)).Elem(),
		"Request":      http.Request{},
		"Response":     http.Response{},
		"RoundTripper": reflect.TypeOf((*http.RoundTripper)(nil)).Elem(),
	})
}
