package main

import (
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/api/users", apiUsers)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/users", users)
}
