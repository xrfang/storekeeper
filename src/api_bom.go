package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"storekeeper/db"
)

func apiBom(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[9:])
	switch r.Method {
	case "GET":
		so, _ := strconv.Atoi(r.URL.Query().Get("order"))
		var res map[string]interface{}
		bill, items, err := db.GetBill(id, so)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			panic(err)
		}
		users, err := db.ListUsers(1)
		assert(err)
		res = map[string]interface{}{"bill": bill, "items": items}
		for _, u := range users {
			if u.ID == bill.User {
				res["user"] = u.Name
				break
			}
		}
		w.Header().Set("Content-Type", "application/json")
		assert(json.NewEncoder(w).Encode(res))
	case "POST":
		assert(r.ParseForm())
		item := r.FormValue("item")
		cnt, _ := strconv.Atoi(r.FormValue("count"))
		if cnt < 0 {
			panic(fmt.Errorf("invalid count"))
		}
		goods, err := db.SearchGoods(item)
		assert(err)
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		if len(items) == 1 {
			fee, _ := strconv.ParseFloat(r.FormValue("fee"), 64)
			memo := r.FormValue("memo")
			bill := db.Bill{ID: id, User: uid, Type: 1, Memo: memo, Fee: fee}
			id, err = db.AddGoodsToBill(bill, goods[0].ID, goods[0].Name, cnt)
			if err != nil {
				if err != db.ErrItemAlreadyExists {
					panic(err)
				}
				cnt = -cnt
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    id,
			"item":  items,
			"count": cnt,
		})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
