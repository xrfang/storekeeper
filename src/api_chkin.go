package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
			bis, err := db.GetBillItems(id, item)
			assert(err)
			w.Header().Set("Content-Type", "application/json")
			assert(json.NewEncoder(w).Encode(bis[0]))
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
			items, err := db.GetBillItems(id, gid)
			assert(err)
			item := items[0]
			cost, err := strconv.ParseFloat(r.FormValue("cost"), 64)
			if err == nil {
				item.Cost = cost
			}
			req, err := strconv.Atoi(r.FormValue("request"))
			if err == nil {
				item.Request = req
			}
			cfm, err := strconv.Atoi(r.FormValue("confirm"))
			if err == nil {
				item.Confirm = cfm
			}
			assert(db.SetBillItem(item, 1))
			return
		}
		memo := r.FormValue("memo")
		if len(memo) > 0 { //编辑单据备注
			//TODO
			return
		}
		item := r.FormValue("item")
		goods, err := db.SearchGoods(item)
		assert(err)
		cfm, _ := strconv.Atoi(r.FormValue("confirm"))
		if cfm > 0 { //验货入库
			var items []interface{}
			for _, g := range goods {
				items = append(items, g.ID)
			}
			if len(items) > 0 {
				//TODO
			}
			return
		}
		//增加新条目
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		req, _ := strconv.Atoi(r.FormValue("request"))
		if len(items) == 1 {
			if id == 0 {
				id, err = db.SetBill(db.Bill{ID: id, User: uid, Type: 1})
				assert(err)
			}
			err = db.SetBillItem(db.BillItem{
				BomID:     id,
				GoodsID:   goods[0].ID,
				GoodsName: goods[0].Name,
				Request:   req,
			}, 0)
			if err != nil {
				if err != db.ErrItemAlreadyExists {
					panic(err)
				}
				req = -req
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    id,
			"item":  items,
			"count": req,
		})
	case "DELETE":
		ids := strings.SplitN(strings.TrimSpace(r.URL.Path[11:]), "/", 2)
		bid, _ := strconv.Atoi(ids[0])
		if bid <= 0 {
			panic(errors.New("invalid bom id"))
		}
		gid := 0
		if len(ids) == 2 && len(ids[1]) > 0 {
			gid, _ = strconv.Atoi(ids[1])
			if gid <= 0 {
				panic(errors.New("invalid goods id"))
			}
		}
		if gid == 0 {
			assert(db.DeleteBill(bid))
		} else {
			assert(db.DeleteBillItem(bid, gid))
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
