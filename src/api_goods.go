package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"storekeeper/db"
)

func apiGoods(w http.ResponseWriter, r *http.Request) {
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	switch r.Method {
	case "GET":
		bt, _ := strconv.Atoi(r.URL.Query().Get("type"))
		if bt != 1 && bt != 2 {
			http.Error(w, "missing `type`", http.StatusBadRequest)
			return
		}
		res, err := db.SearchGoods(r.URL.Query().Get("terms"), bt)
		assert(err)
		w.Header().Set("Content-Type", "application/json")
		assert(json.NewEncoder(w).Encode(res))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
