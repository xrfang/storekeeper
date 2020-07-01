package main

import (
	"net/http"
	"storekeeper/db"
	"strconv"
)

func invChkList(w http.ResponseWriter, r *http.Request) {
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
	db.RemoveEmptyBills()
	renderTemplate(w, "invchk.html", nil)
}

func invChkEdit(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[8:])
	switch r.Method {
	case "GET":
		u := db.GetUser(uid)
		if id == 0 {
			id = db.InventoryWIP()
			if id == 0 {
				id = db.SetBill(db.Bill{Type: 3, User: uid})
			}
			db.UpdateInventory(id)
			id = -id
		} else {
			db.UpdateInventory(id)
		}
		renderTemplate(w, "invchked.html", map[string]interface{}{"user": u, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func invChkEditItem(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[13:])
	switch r.Method {
	case "POST":
		assert(r.ParseForm())
		item := r.FormValue("item")
		goods := db.SearchGoods(item)
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		req := 0
		cfm, _ := strconv.Atoi(r.FormValue("confirm"))
		if cfm > 0 && len(items) == 1 {
			bis := db.GetBillItems(id, goods[0].ID)
			if len(bis) == 0 {
				bi := db.BillItem{
					BomID:     id,
					GoodsID:   goods[0].ID,
					GoodsName: goods[0].Name,
					Request:   0,
					Confirm:   cfm,
				}
				db.SetBillItem(bi, 0)
			} else {
				req = bis[0].Request
				bis[0].Confirm = cfm
				db.SetBillItem(bis[0], 1)
			}
		}
		jsonReply(w, map[string]interface{}{
			"id":   id,
			"item": items,
			"req":  req,
			"cfm":  cfm,
		})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
