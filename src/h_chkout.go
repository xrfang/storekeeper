package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func chkOutList(w http.ResponseWriter, r *http.Request) {
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
	assert(db.RemoveEmptyBills())
	bills, err := db.ListBills(&db.Bill{Type: 2})
	assert(err)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	users, err := db.ListUsers(1)
	assert(err)
	renderTemplate(w, "chkout.html", map[string]interface{}{"bills": bm, "users": users})
}

func chkOutEdit(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[8:])
	switch r.Method {
	case "GET":
		var us []db.User
		if id == 0 {
			var err error
			us, err = db.ListUsers(uid)
			assert(err)
			id, err = db.SetBill(db.Bill{Type: 2, User: uid, Markup: 20})
			assert(err)
			id = -id
		} else {
			b, _, err := db.GetBill(id, -1)
			assert(err)
			u, err := db.GetUser(b.User)
			assert(err)
			if u.Client == 0 {
				us, err = db.ListUsers(u.ID)
			} else {
				us, err = db.ListUsers(u.Client)
			}
			assert(err)
		}
		renderTemplate(w, "chkouted.html", map[string]interface{}{"users": us, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func chkOutSetFee(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[12:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	json.NewEncoder(w).Encode(r.Form)
	fee, err := strconv.ParseFloat(r.FormValue("fee"), 64)
	assert(err)
	b, _, err := db.GetBill(id, -1)
	assert(err)
	b.Fee = fee
	_, err = db.SetBill(b)
	assert(err)
}

func chkOutSetMarkup(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	json.NewEncoder(w).Encode(r.Form)
	markup, err := strconv.Atoi(r.FormValue("markup"))
	assert(err)
	b, _, err := db.GetBill(id, -1)
	assert(err)
	b.Markup = markup
	_, err = db.SetBill(b)
	assert(err)
}

func chkOutSetRequester(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[12:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	json.NewEncoder(w).Encode(r.Form)
	req, _ := strconv.Atoi(r.FormValue("user"))
	if req <= 0 {
		panic(fmt.Errorf("invalid user_id"))
	}
	b, _, err := db.GetBill(id, -1)
	assert(err)
	b.User = req
	_, err = db.SetBill(b)
	assert(err)
}
