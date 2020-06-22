package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
	"strings"
)

func apiSetProp(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[9:])
	if id <= 0 {
		panic(fmt.Errorf("invalid ID"))
	}
	assert(r.ParseForm())
	key := r.FormValue("key")
	val := r.FormValue("val")
	switch key {
	case "amount":
		v := strings.SplitN(val, ",", 2)
		gid, _ := strconv.Atoi(v[0])
		amt, _ := strconv.Atoi(v[1])
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
	case "stat":
		stat, _ := strconv.Atoi(val)
		db.SetInventoryByBill(id, stat)
	case "memo":
		b, _ := db.GetBill(id, -1)
		b.Memo = val
		db.SetBill(b)
	case "fee":
		fee, err := strconv.ParseFloat(val, 64)
		assert(err)
		b, _ := db.GetBill(id, -1)
		b.Fee = fee
		db.SetBill(b)
	case "markup":
		markup, err := strconv.Atoi(val)
		assert(err)
		b, _ := db.GetBill(id, -1)
		b.Markup = markup
		db.SetBill(b)
	case "user":
		u := db.GetUser(val)
		b, _ := db.GetBill(id, -1)
		b.User = u.ID
		db.SetBill(b)
	case "sets":
		sets, _ := strconv.Atoi(val)
		if sets <= 0 {
			panic(fmt.Errorf("invalid sets"))
		}
		b, _ := db.GetBill(id, -1)
		b.Sets = sets
		db.SetBill(b)
	case "paid":
		pay, err := strconv.ParseFloat(val, 64)
		if err != nil {
			panic(err)
		}
		b, _ := db.GetBill(id, -1)
		b.Paid = pay
		b.Status = 2
		db.SetBill(b)
	}
}
