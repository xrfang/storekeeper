package main

import (
	"net/http"
	"storekeeper/db"
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
