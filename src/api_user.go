package main

import (
	"encoding/json"
	"net/http"

	"storekeeper/db"
)

func apiUsers(w http.ResponseWriter, r *http.Request) {
	if !validate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case "GET":
		users, err := db.ListUsers(1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for i := 0; i < 20; i++ {
			u := users[0]
			if i%2 != 0 {
				u.Client = 1
			}
			users = append(users, u)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
