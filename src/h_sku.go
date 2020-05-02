package main

import (
	"net/http"
)

func sku(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	_ = uid //TODO
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		renderTemplate(w, "sku.html", nil)
	case "POST":
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}