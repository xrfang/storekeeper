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
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	db.RemoveEmptyBills()
	month := r.URL.Query().Get("month")
	user := r.URL.Query().Get("user")
	renderTemplate(w, "chkout.html", struct {
		Month string
		User  string
	}{month, user})
}

func chkOutEdit(w http.ResponseWriter, r *http.Request) {
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
		var us []db.User
		if id == 0 {
			us = db.ListUsers(uid)
			u := db.GetUser(uid)
			m := cf.Markup
			if u.Markup >= 0 {
				m = u.Markup
			}
			id = db.SetBill(db.Bill{Type: 2, User: uid, Sets: 1, Markup: m, Fee: 0})
			id = -id
		} else {
			b, _ := db.GetBill(id, -1)
			ref, _ := strconv.Atoi(r.URL.Query().Get("ref"))
			if ref > 0 {
				db.CloneBillItems(b, ref)
			}
			u := db.GetUser(b.User)
			if uid == 1 {
				us = db.ListUsers(1)
			} else if u.Client == 0 {
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

func chkOutEditItem(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[13:])
	switch r.Method {
	case "POST":
		assert(r.ParseForm())
		mode, _ := strconv.Atoi(r.FormValue("mode"))
		fmt.Println("TODO: check mode...")
		rx := r.FormValue("rx")
		res := make(map[string]interface{})
		ps := db.GetPSItems(rx)
		if db.IsBillEmpty(id) && len(ps) > 1 {
			res["reference"] = db.GetPrevRx(ps)
		}
		for i, p := range ps {
			if len(p.Items) == 1 && p.Weight > 0 {
				if db.SetBillItem(db.BillItem{
					BomID:     id,
					Cost:      p.Items[0].Cost,
					GoodsID:   p.Items[0].ID,
					GoodsName: p.Items[0].Name,
					Memo:      p.Memo,
					Request:   p.Weight,
				}, 0) {
					ps[i].Weight = -p.Weight
				}
			}
		}
		res["rx_items"] = ps
		jsonReply(w, res)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
