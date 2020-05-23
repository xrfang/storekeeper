package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func invChkList(w http.ResponseWriter, r *http.Request) {
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	assert(db.RemoveEmptyBills())
	bills, err := db.ListBills(&db.Bill{Type: 3})
	assert(err)
	fmt.Printf("invchk: %+v\n", bills)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	u, err := db.GetUser(uid)
	assert(err)
	renderTemplate(w, "invchk.html", map[string]interface{}{"bills": bm, "user": u})
}

func invChkEdit(w http.ResponseWriter, r *http.Request) {
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
		renderTemplate(w, "invchked.html", map[string]interface{}{"user": u, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
