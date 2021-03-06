package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"storekeeper/db"
)

func apiSkuFind(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	jsonReply(w, db.FindSKU(r.URL.Query().Get("py")))
}

func apiSkuSearch(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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
