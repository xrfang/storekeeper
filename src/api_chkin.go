package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"storekeeper/db"
)

func apiChkIn(w http.ResponseWriter, r *http.Request) {
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
	switch r.Method {
	case "GET":
		item, _ := strconv.Atoi(r.URL.Query().Get("item"))
		if item > 0 { //在弹出框中编辑条目
			bi, err := db.GetBillItem(id, item)
			assert(err)
			w.Header().Set("Content-Type", "application/json")
			assert(json.NewEncoder(w).Encode(bi))
			return
		}
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
		w.Header().Set("Content-Type", "application/json")
		assert(json.NewEncoder(w).Encode(res))
	case "POST":
		assert(r.ParseForm())
		gid, _ := strconv.Atoi(r.FormValue("gid"))
		if gid > 0 { //提交单条目编辑
			//TODO
			return
		}
		memo := r.FormValue("memo")
		if len(memo) > 0 { //编辑单据备注
			//TODO
			return
		}
		//增加新条目
		item := r.FormValue("item")
		cnt, _ := strconv.Atoi(r.FormValue("count"))
		goods, err := db.SearchGoods(item)
		assert(err)
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		if len(items) == 1 {
			if id == 0 {
				id, err = db.SetBill(db.Bill{ID: id, User: uid, Type: 1})
				assert(err)
			}
			err = db.SetBillItem(db.BillItem{
				BomID:     id,
				GoodsID:   goods[0].ID,
				GoodsName: goods[0].Name,
				Request:   cnt,
			}, 0)
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