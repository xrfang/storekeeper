package main

import (
	"net/http"
)

func inventory(w http.ResponseWriter, r *http.Request) {
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	renderTemplate(w, "inventory.html", nil)
}
