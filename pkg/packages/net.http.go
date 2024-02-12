//go:build !appengine

package packages

import (
	"net/http"
)

func init() {
	Packages["net/http"] = map[string]any{
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
	}
	PackageTypes["net/http"] = map[string]any{
		"Client":   http.Client{},
		"Cookie":   http.Cookie{},
		"Request":  http.Request{},
		"Response": http.Response{},
	}
}
