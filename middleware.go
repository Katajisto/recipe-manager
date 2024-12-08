package main

import (
	"net/http"
	"strings"
)

func ConfigMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := GetConfig()

		if conf == nil && !strings.Contains(r.URL.Path, "static") && !strings.Contains(r.URL.Path, "doinit") {
			views.ExecuteTemplate(w, "init.tmpl", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
