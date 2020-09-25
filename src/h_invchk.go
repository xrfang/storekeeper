package main

import (
	"fmt"
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
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[8:])
	if id == 0 {
		id = db.InventoryWIP()
	}
	if id == 0 {
		id = db.CreateInventory(uid)
	}
	u := db.GetUser(uid)
	renderTemplate(w, "invchked.html", map[string]interface{}{"user": u, "bill": id})
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
		ps := db.GetPSItems(item)
		if len(ps) == 0 {
			http.Error(w, "输入错误", http.StatusBadRequest)
			return
		}
		errItems := db.PSItems{}
		for _, p := range ps {
			if len(p.Items) != 1 {
				errItems = append(errItems, p)
				continue
			}
			bis := db.GetBillItems(id, p.Items[0].ID)
			if len(bis) != 1 {
				panic(fmt.Errorf("GetBillItems failed: bid=%v; gid=%v", id, p.Items[0].ID))
			}
			bis[0].Confirm = p.Weight
			bis[0].Flag = 1
			db.SetBillItem(bis[0], 1)
		}
		jsonReply(w, errItems)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
