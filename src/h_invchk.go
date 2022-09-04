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
		pick := db.PSItems{}
		done := db.PSItems{}
		for _, p := range ps {
			if len(p.Items) != 1 {
				pick = append(pick, p)
				continue
			}
			db.SetBillItem(db.BillItem{
				BomID:     id,
				Cost:      p.Items[0].Cost,
				GoodsID:   p.Items[0].ID,
				GoodsName: p.Items[0].Name,
				Memo:      p.Memo,
				Request:   *p.Weight,
			}, 2) //盘点单的mode=2
			if p.Rack != "" {
				//TODO: 盘点时修改Rack其实是不妥的，因为录入以后才知道应当放在哪个
				//Rack中。分多次录入就会导致混乱！
				//db.UpdateRack(p.Items[0].ID, p.Rack)
			}
			done = append(done, p)
		}
		jsonReply(w, map[string]interface{}{"done": done, "pick": pick})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
