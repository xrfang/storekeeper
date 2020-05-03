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
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		cnt, err := db.CountSKU()
		assert(err)
		renderTemplate(w, "sku.html", struct{ Total int }{cnt})
	case "POST":
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}