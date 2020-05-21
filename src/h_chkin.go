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
	assert(db.RemoveEmptyBills())
	bills, err := db.ListBills(&db.Bill{Type: 1})
	assert(err)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	users, err := db.ListUsers(1)
	assert(err)
	renderTemplate(w, "chkin.html", map[string]interface{}{"bills": bm, "users": users})
}

func chkInEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, trace("%v", e).Error(), http.StatusInternalServerError)
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
		if id == 0 {
			id, err = db.SetBill(db.Bill{Type: 1, User: uid})
			assert(err)
			id = -id
		}
		renderTemplate(w, "chkined.html", map[string]interface{}{"user": u, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func chkInSetMemo(w http.ResponseWriter, r *http.Request) {
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
	assert(r.ParseForm())
	memo := r.FormValue("memo")
	b, _, err := db.GetBill(id, -1)
	assert(err)
	b.Memo = memo
	_, err = db.SetBill(b)
	assert(err)
}
