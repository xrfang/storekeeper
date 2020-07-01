package main

import (
	"net/http"

	"storekeeper/db"
)

func apiUsers(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, trace("%v", e).Error(), http.StatusInternalServerError)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case "GET":
		users := db.ListUsers(uid)
		jsonReply(w, users)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
