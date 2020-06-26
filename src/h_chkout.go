package main

import (
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
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	db.RemoveEmptyBills()
	month := r.URL.Query().Get("month")
	renderTemplate(w, "chkout.html", month)
}

func chkOutEdit(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
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
			id = db.SetBill(db.Bill{Type: 2, User: uid, Sets: 1, Markup: 30, Fee: 0})
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

func chkOutEditItem(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
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
