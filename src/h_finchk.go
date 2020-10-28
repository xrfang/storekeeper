package main

import (
	"net/http"
	"storekeeper/db"
	"strconv"
)

func finChkList(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	renderTemplate(w, "finchk.html", nil)
}

func finChkEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[8:])
	//TODO：以下代码是从invchk中copy过来的，需要修改，并创建finchked.html
	if id == 0 {
		id = db.InventoryWIP()
	}
	if id == 0 {
		id = db.CreateInventory(uid)
	}
	u := db.GetUser(uid)
	renderTemplate(w, "finchked.html", map[string]interface{}{"user": u, "bill": id})
}
