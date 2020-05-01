package main

import (
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/api/users", apiUsers)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/otp/", otp)
	http.HandleFunc("/users", users)
	http.HandleFunc("/users/", users)
}
