package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"storekeeper/db"
)

func apiSkuFind(w http.ResponseWriter, r *http.Request) {
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
	jsonReply(w, db.FindSKU(r.URL.Query().Get("py")))
}

func apiSkuSearch(w http.ResponseWriter, r *http.Request) {
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
		qr := db.QuerySKU(terms)
		jsonReply(w, qr)
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
		id, _ := strconv.Atoi(r.URL.Path[9:])
		goods := db.GetSKUs(id)
		if len(goods) == 0 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		jsonReply(w, goods[0])
	case "POST":
		var skus []db.Goods
		assert(json.NewDecoder(r.Body).Decode(&skus))
		db.UpdateSKUs(skus)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
