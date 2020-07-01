package main

import (
	"net/http"
	"storekeeper/db"
)

func logout(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	err := db.Logout(getCookie(r, "token"))
	if err != nil {
		http.Error(w, "Logout: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
