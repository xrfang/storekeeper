package main

import (
	"net/http"

	"storekeeper/db"
)

func sku(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		cnt := db.CountSKU()
		renderTemplate(w, "sku.html", struct{ Total int }{cnt})
	case "POST":
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
