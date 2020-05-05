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
	bills, err := db.GetBills(nil)
	assert(err)
	for _, b := range bills {
		//TODO：根据状态分类
		_ = b
	}
	renderTemplate(w, "chkin.html", nil)
}
