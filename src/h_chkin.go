package main

import (
	"net/http"

	"storekeeper/db"
)

func chkIn(w http.ResponseWriter, r *http.Request) {
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
	bills, err := db.GetBills(&db.Bill{Type: 1})
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
	renderTemplate(w, "chkin.html", bm)
}
