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
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
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
		v := strings.SplitN(val, ",", 3)
		gid, _ := strconv.Atoi(v[0])
		amt, _ := float(v[1])
		ext, _ := strconv.Atoi(v[2])
		db.GetBill(id, -1)
		if amt > 0 {
			bis := db.GetBillItems(id, gid)
			db.SetBillItem(db.BillItem{
				BomID:     id,
				Cost:      bis[0].Cost,
				GoodsName: bis[0].GoodsName,
				GoodsID:   gid,
				Request:   amt,
				Flag:      ext,
				Memo:      bis[0].Memo,
			}, 1)
		} else {
			db.DeleteBillItem(id, gid)
		}
	case "itememo":
		v := strings.SplitN(val, ",", 2)
		gid, _ := strconv.Atoi(v[0])
		memo := v[1]
		db.GetBill(id, -1)
		bis := db.GetBillItems(id, gid)
		db.SetBillItem(db.BillItem{
			BomID:     id,
			Cost:      bis[0].Cost,
			GoodsName: bis[0].GoodsName,
			GoodsID:   gid,
			Request:   bis[0].Request,
			Flag:      bis[0].Flag,
			Memo:      memo,
		}, 1)
	case "setinv":
		db.SetInventoryByBill(id)
	case "memo":
		b, _ := db.GetBill(id, -1)
		b.Memo = val
		db.SetBill(b)
	case "fee":
		fee, err := float(val)
		assert(err)
		b, _ := db.GetBill(id, -1)
		b.Fee = fee
		db.SetBill(b)
	case "markup":
		markup, err := float(val)
		assert(err)
		b, _ := db.GetBill(id, -1)
		b.Markup = markup
		u := db.GetUser(b.User)
		u.Markup = markup
		db.SetBill(b)
		db.UpdateUser(u)
	case "user":
		u := db.GetUser(val)
		b, _ := db.GetBill(id, -1)
		b.User = u.ID
		if u.Markup < 0 {
			b.Markup = cf.Markup
		} else {
			b.Markup = u.Markup
		}
		db.SetBill(b)
	case "sets":
		sets, _ := strconv.Atoi(val)
		if sets <= 0 {
			panic(fmt.Errorf("invalid sets"))
		}
		b, _ := db.GetBill(id, -1)
		b.Sets = sets
		db.SetBill(b)
	case "ship":
		b, _ := db.GetBill(id, -1)
		if b.Status < 1 || b.Status > 2 {
			panic(fmt.Errorf("bill #%d: expect status 1 or 2 when shipping", id))
		}
		b.Courier = val
		b.Status = 2
		db.SetBill(b)
	case "paid":
		pay, err := float(val)
		if err != nil {
			panic(err)
		}
		b, _ := db.GetBill(id, -1)
		if b.Status != 2 {
			panic(fmt.Errorf("bill #%d: expect status 2 when paying", id))
		}
		b.Paid = pay
		b.Status = 3
		db.SetBill(b)
	}
}
