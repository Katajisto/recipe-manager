package main

import (
	"net/http"
)

func ConfigMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := GetConfig()

		if conf == nil {
			views.ExecuteTemplate(w, "init.tmpl", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("auth")
		// TODO: this is not safe for sure
		if err != nil {
			views.ExecuteTemplate(w, "login.tmpl", "")
			return
		}

		if !CheckToken(authCookie.Value) {
			http.Error(w, "Your token is invalid", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
