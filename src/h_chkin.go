package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"storekeeper/db"
)

func chkInList(w http.ResponseWriter, r *http.Request) {
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
	bills, err := db.ListBills(&db.Bill{Type: 1})
	assert(err)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	for i := 1; i < 5; i++ {
		_, ok := bm[byte(i)]
		if !ok {
			bm[byte(i)] = nil
		}
	}
	users, err := db.ListUsers(1)
	assert(err)
	renderTemplate(w, "chkin.html", struct {
		Bills map[byte][]db.Bill
		Users []db.User
	}{bm, users})
}

func chkInEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[7:])
	switch r.Method {
	case "GET":
		u, err := db.GetUser(uid)
		assert(err)
		var bill *db.Bill
		var items []db.BillItem
		if id != 0 {
			bill, items, err = db.GetBill(id)
			assert(err)
		} else {
			bill = &db.Bill{ID: 0, Status: 1}
		}
		renderTemplate(w, "chkined.html", struct {
			User  *db.User
			Bill  *db.Bill
			Items []db.BillItem
		}{u, bill, items})
	case "POST":
		assert(r.ParseForm())
		item := r.FormValue("item")
		cnt, _ := strconv.Atoi(r.FormValue("count"))
		if cnt <= 0 {
			panic(fmt.Errorf("invalid count"))
		}
		goods, err := db.SearchGoods(item)
		assert(err)
		var items []string
		for _, g := range goods {
			items = append(items, g.Name)
		}
		if len(items) == 1 {
			if id == 0 {

			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    id,
			"item":  items,
			"count": cnt,
		})
	}
}
