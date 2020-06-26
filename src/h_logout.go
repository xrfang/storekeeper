package main

import (
	"net/http"
	"storekeeper/db"
)

func logout(w http.ResponseWriter, r *http.Request) {
	err := db.Logout(getCookie(r, "token"))
	if err != nil {
		http.Error(w, "Logout: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
