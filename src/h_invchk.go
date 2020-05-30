package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func invChkList(w http.ResponseWriter, r *http.Request) {
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	db.RemoveEmptyBills()
	bills := db.ListBills(&db.Bill{Type: 3})
	fmt.Printf("invchk: %+v\n", bills)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	u := db.GetUser(uid)
	renderTemplate(w, "invchk.html", map[string]interface{}{"bills": bm, "user": u})
}

func invChkEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, trace("%v", e).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[8:])
	switch r.Method {
	case "GET":
		u := db.GetUser(uid)
		if id == 0 {
			bills := db.ListBills(&db.Bill{Type: 3, Status: 0})
			if len(bills) > 0 {
				http.Error(w, "只允许一个进行中盘点", http.StatusBadRequest)
				return
			}
			id = db.SetBill(db.Bill{Type: 3, User: uid})
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
			http.Error(w, trace("%v", e).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
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
