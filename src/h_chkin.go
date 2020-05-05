package main

import (
	"net/http"

	"storekeeper/db"
)

func chkIn(w http.ResponseWriter, r *http.Request) {
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	bills := db.GetBills(nil)
	new := 0
	for _, b := range bills {
		//TODO：根据状态分类
	}
	renderTemplate(w, "chkin.html", nil)
}
