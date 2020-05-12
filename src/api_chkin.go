package main

import (
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
			jsonReply(w, bis[0])
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
		jsonReply(w, res)
	case "POST":
		assert(r.ParseForm())
		status, _ := strconv.Atoi(r.FormValue("status"))
		if status > 0 {
			b, _, err := db.GetBill(id, -1) //第二参数不是0或1表示不需要获取条目
			assert(err)
			b.Status = 1
			_, err = db.SetBill(b)
			assert(err)
			return
		}
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
		if cfm != 0 { //验货入库
			res := map[string]interface{}{"id": id, "item": []string{}, "count": cfm}
			var items []interface{}
			for _, g := range goods {
				items = append(items, g.ID)
			}
			if len(items) == 0 {
				jsonReply(w, res)
				return
			}
			bis, err := db.GetBillItems(id, items...)
			assert(err)
			switch len(bis) {
			case 0:
				jsonReply(w, res)
			case 1:
				res["item"] = []string{bis[0].GoodsName}
				diff := bis[0].Request - bis[0].Confirm
				if cfm > diff {
					res["count"] = diff - cfm
					jsonReply(w, res)
					return
				}
				bis[0].Confirm += cfm
				assert(db.SetBillItem(bis[0], 1))
				res["count"] = bis[0].Confirm
				jsonReply(w, res)
			default:
				var names []string
				for _, bi := range bis {
					names = append(names, bi.GoodsName)
				}
				res["item"] = names
				jsonReply(w, res)
			}
			return
		}
		//增加新条目
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		req, _ := strconv.Atoi(r.FormValue("request"))
		if req > 0 && len(items) == 1 {
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
		jsonReply(w, map[string]interface{}{
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
