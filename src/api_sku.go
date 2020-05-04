package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"storekeeper/db"
)

func apiSkuList(w http.ResponseWriter, r *http.Request) {
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
		tm := make(map[string]bool)
		for _, t := range strings.Split(r.URL.Query().Get("terms"), " ") {
			t = strings.TrimSpace(t)
			if len(t) > 0 {
				tm[t] = true
			}
		}
		var terms []string
		for t := range tm {
			terms = append(terms, t)
		}
		qr, err := db.QuerySKU(terms)
		assert(err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(qr)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func apiSkuEdit(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	case "POST":
		var skus []db.Goods
		assert(json.NewDecoder(r.Body).Decode(&skus))
		db.UpdateSKUs(skus)
	}
}
