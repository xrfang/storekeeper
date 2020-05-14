package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func apiChkOut(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, trace("%v", e).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[11:])
	fmt.Println("apiChkOut, id:", id)
	switch r.Method {
	case "GET":
		//TODO：条目编辑
		/*
			item, _ := strconv.Atoi(r.URL.Query().Get("item"))
			if item > 0 { //在弹出框中编辑条目
				bis, err := db.GetBillItems(id, item)
				assert(err)
				jsonReply(w, bis[0])
				return
			}
		*/
		//编辑整个单据
		so, _ := strconv.Atoi(r.URL.Query().Get("order"))
		var (
			res   map[string]interface{}
			bill  db.Bill
			items []db.BillItem
			err   error
		)
		if id == 0 {
			bill = db.Bill{User: uid}
		} else {
			bill, items, err = db.GetBill(id, so)
			assert(err)
		}
		res = map[string]interface{}{"bill": bill, "items": items}
		users, err := db.ListUsers(1)
		assert(err)
		for _, u := range users {
			if u.ID == bill.User {
				res["user"] = u.Name
				break
			}
		}
		jsonReply(w, res)
	case "POST":
		assert(r.ParseForm())
	case "DELETE":
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
