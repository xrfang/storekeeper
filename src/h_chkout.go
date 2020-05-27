package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func chkOutList(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	db.RemoveEmptyBills()
	bills := db.ListBills(&db.Bill{Type: 2})
	bm := make(map[byte][]db.Bill)
	for _, b := range bills {
		bm[b.Status] = append(bm[b.Status], b)
	}
	users := db.ListUsers(1)
	renderTemplate(w, "chkout.html", map[string]interface{}{"bills": bm, "users": users})
}

func chkOutBill(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	so, _ := strconv.Atoi(r.URL.Query().Get("order"))
	var (
		res   map[string]interface{}
		bill  db.Bill
		items []db.BillItem
	)
	bill, items = db.GetBill(id, so)
	res = map[string]interface{}{"bill": bill, "items": items}
	jsonReply(w, res)
}

func chkOutEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
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
		var us []db.User
		if id == 0 {
			us = db.ListUsers(uid)
			id = db.SetBill(db.Bill{Type: 2, User: uid, Sets: 1, Markup: 20, Fee: 0})
			id = -id
		} else {
			b, _ := db.GetBill(id, -1)
			u := db.GetUser(b.User)
			if u.Client == 0 {
				us = db.ListUsers(u.ID)
			} else {
				us = db.ListUsers(u.Client)
			}
		}
		renderTemplate(w, "chkouted.html", map[string]interface{}{"users": us, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func chkOutSetFee(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[12:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	fee, err := strconv.ParseFloat(r.FormValue("fee"), 64)
	assert(err)
	b, _ := db.GetBill(id, -1)
	b.Fee = fee
	db.SetBill(b)
}

func chkOutSetMarkup(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	markup, err := strconv.Atoi(r.FormValue("markup"))
	assert(err)
	b, _ := db.GetBill(id, -1)
	b.Markup = markup
	db.SetBill(b)
}

func chkOutSetRequester(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[12:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	req, _ := strconv.Atoi(r.FormValue("user"))
	if req <= 0 {
		panic(fmt.Errorf("invalid user_id"))
	}
	b, _ := db.GetBill(id, -1)
	b.User = req
	db.SetBill(b)
}

func chkOutSetSets(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	sets, _ := strconv.Atoi(r.FormValue("sets"))
	if sets <= 0 {
		panic(fmt.Errorf("invalid sets"))
	}
	b, _ := db.GetBill(id, -1)
	b.Sets = sets
	db.SetBill(b)
}

func chkOutEditItem(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		req, _ := strconv.Atoi(r.FormValue("request"))
		if req > 0 && len(items) == 1 {
			if db.SetBillItem(db.BillItem{
				BomID:     id,
				Cost:      goods[0].Cost,
				GoodsID:   goods[0].ID,
				GoodsName: goods[0].Name,
				Request:   req,
			}, 0) {
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

func chkOutSetAmount(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[12:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	gid, _ := strconv.Atoi(r.FormValue("gid"))
	amt, _ := strconv.Atoi(r.FormValue("amt"))
	db.GetBill(id, -1)
	if amt > 0 {
		bis := db.GetBillItems(id, gid)
		db.SetBillItem(db.BillItem{
			BomID:     id,
			Cost:      bis[0].Cost,
			GoodsName: bis[0].GoodsName,
			GoodsID:   gid,
			Request:   amt,
		}, 1)
	} else {
		db.DeleteBillItem(id, gid)
	}
}

func chkOutSetStat(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	stat, _ := strconv.Atoi(r.FormValue("stat"))
	db.SetInventoryByBill(id, stat)
}

func chkOutSetMemo(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[13:])
	assert(r.ParseForm())
	memo := r.FormValue("memo")
	b, _ := db.GetBill(id, -1)
	b.Memo = memo
	db.SetBill(b)
}
