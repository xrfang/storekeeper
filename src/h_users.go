package main

import (
	"net/http"
)

func users(w http.ResponseWriter, r *http.Request) {
	if !validate(r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	renderTemplate(w, "users.html", nil)
}
