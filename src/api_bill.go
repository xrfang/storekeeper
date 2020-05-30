package main

import (
	"fmt"
	"net/http"
	"strconv"

	"storekeeper/db"
)

func apiGetBill(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[9:])
	if id < 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	so, _ := strconv.Atoi(r.URL.Query().Get("order"))
	var (
		res   map[string]interface{}
		bill  db.Bill
		items []db.BillItem
	)
	if id == 0 {
		bill = db.Bill{User: uid}
	} else {
		bill, items = db.GetBill(id, so)
	}
	res = map[string]interface{}{"bill": bill, "items": items}
	users := db.ListUsers(1)
	for _, u := range users {
		if u.ID == bill.User {
			res["user"] = u.Name
			break
		}
	}
	jsonReply(w, res)
}
