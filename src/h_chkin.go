package main

import (
	"net/http"
	"strconv"

	"storekeeper/db"
)

func chkInList(w http.ResponseWriter, r *http.Request) {
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
	renderTemplate(w, "chkin.html", nil)
}

func chkInEdit(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[7:])
	switch r.Method {
	case "GET":
		u := db.GetUser(uid)
		if id == 0 {
			id = db.SetBill(db.Bill{Type: 1, User: uid})
			id = -id
		}
		renderTemplate(w, "chkined.html", map[string]interface{}{"user": u, "bill": id})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func chkInSetMemo(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[12:])
	assert(r.ParseForm())
	memo := r.FormValue("memo")
	b, _ := db.GetBill(id, -1)
	b.Memo = memo
	db.SetBill(b)
}

func chkInEditItem(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[12:])
	switch r.Method {
	case "POST":
		assert(r.ParseForm())
		_, nu := db.AnalyzeGoodsUsage()
		stock := make(map[string]int)
		for _, u := range nu {
			stock[u.Name] = u.Amount
		}
		rx := r.FormValue("rx")
		res := make(map[string]interface{})
		ps := db.GetPSItems(rx)
		for _, p := range ps {
			if len(p.Items) == 1 && p.Weight > 0 {
				db.SetBillItem(db.BillItem{
					BomID:     id,
					Cost:      p.Items[0].Cost,
					GoodsID:   p.Items[0].ID,
					GoodsName: p.Items[0].Name,
					Memo:      p.Memo,
					Request:   p.Weight,
				}, 0)
			}
		}
		var unused []db.UsageInfo
		_, items := db.GetBill(id, 0)
		for _, it := range items {
			amt := stock[it.GoodsName]
			if amt > 0 {
				unused = append(unused, db.UsageInfo{
					Name:   it.GoodsName,
					Amount: amt,
					Batch:  1,
				})
			}

		}
		res["rx_items"] = ps
		res["unused"] = unused
		jsonReply(w, res)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
