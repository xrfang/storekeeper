package main

import (
	"fmt"
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
	if db.InventoryWIP() != 0 {
		http.Error(w, "当前有未结束的盘点", http.StatusConflict)
		return
	}
	//_probe参数表示客户端使用jquery测试本页面是否可以跳转，如果盘点进行中则不允许访问
	if r.URL.Query().Get("_probe") != "" {
		fmt.Fprintln(w, "OK") //可以继续访问
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
		res := make(map[string]interface{})
		assert(r.ParseForm())
		rx := r.FormValue("rx")
		ps := db.GetPSItems(rx)
		mode, _ := strconv.Atoi(r.FormValue("mode"))
		switch mode {
		case 0:
			_, nu := db.AnalyzeGoodsUsage()
			stock := make(map[string]int)
			for _, u := range nu {
				stock[u.Name] = u.Amount
			}
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
			res["unused"] = unused
		default:
			_, items := db.GetBill(id, 0)
			itm := make(map[string]*db.BillItem)
			for i, it := range items {
				itm[it.GoodsName] = &items[i]
			}
			for i, p := range ps {
				p.MatchItems(itm)
				if len(p.Items) == 1 && p.Weight != 0 {
					it := itm[p.Items[0].Name]
					it.Confirm += p.Weight
					db.SetBillItem(*it, 1)
				}
				ps[i] = p
			}
		}
		res["rx_items"] = ps
		jsonReply(w, res)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
