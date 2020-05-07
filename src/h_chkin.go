package main

import (
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
}
