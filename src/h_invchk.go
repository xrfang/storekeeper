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
	assert(db.RemoveEmptyBills())
	bills, err := db.ListBills(&db.Bill{Type: 3})
	assert(err)
	fmt.Printf("invchk: %+v\n", bills)
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	u, err := db.GetUser(uid)
	assert(err)
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
	id, _ := strconv.Atoi(r.URL.Path[7:])
	switch r.Method {
	case "GET":
		u, err := db.GetUser(uid)
		assert(err)
		if id == 0 {
			bills, err := db.ListBills(&db.Bill{Type: 3, Status: 0})
			assert(err)
			if len(bills) > 0 {
				http.Error(w, "只允许一个进行中盘点", http.StatusBadRequest)
				return
			}
			id, err = db.SetBill(db.Bill{Type: 1, User: uid})
			assert(err)
			id = -id
			gs, err := db.GetSKUs()
			assert(err)

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
		goods, err := db.SearchGoods(item)
		assert(err)
		items := []string{}
		for _, g := range goods {
			items = append(items, g.Name)
		}
		req, _ := strconv.Atoi(r.FormValue("request"))
		if req > 0 && len(items) == 1 {
			err = db.SetBillItem(db.BillItem{
				BomID:     id,
				Cost:      goods[0].Cost,
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
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
