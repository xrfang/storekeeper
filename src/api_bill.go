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
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
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
		bill  db.Bill
		items []db.BillItem
	)
	if id == 0 {
		bill = db.Bill{User: uid}
	} else {
		bill, items = db.GetBill(id, so)
	}
	res := make(map[string]interface{})
	users := db.ListUsers(1)
	for _, u := range users {
		if u.ID == bill.User {
			if id == 0 {
				if u.Markup < 0 {
					bill.Markup = cf.Markup //系统默认
				} else {
					bill.Markup = u.Markup
				}
			}
			res["user"] = u.Name
			break
		}
	}
	res["bill"] = bill
	res["items"] = items
	jsonReply(w, res)
}
