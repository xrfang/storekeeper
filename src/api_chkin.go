package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"storekeeper/db"
)

func apiChkInList(w http.ResponseWriter, r *http.Request) {
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
	um := make(map[int]string)
	for _, u := range db.ListUsers(1, "id", "name") {
		um[u.ID] = u.Name
	}

	thisMonth := time.Now().Format("2006-01")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = getCookie(r, "bm")
		if month == "" {
			month = thisMonth
		}
	} else {
		setCookie(w, "bm", month, 600)
	}
	summary := db.ListBillSummary(1, 0)
	if len(summary) == 0 || summary[0].Month != thisMonth {
		summary = append([]db.BillSummary{{Month: thisMonth}}, summary...)
	}
	list := db.ListBills(1, 0, month)
	if list == nil {
		list = []db.Bill{}
	}
	jsonReply(w, map[string]interface{}{
		"month":   month,
		"summary": summary,
		"list":    list,
		"users":   um,
	})
}

func apiChkIn(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[11:])
	switch r.Method {
	case "GET":
		item, _ := strconv.Atoi(r.URL.Query().Get("item"))
		if item > 0 { //在弹出框中编辑条目
			bis := db.GetBillItems(id, item)
			jsonReply(w, bis[0])
			return
		}
		//编辑整个单据
		so, _ := strconv.Atoi(r.URL.Query().Get("order"))
		var (
			res   map[string]interface{}
			bill  db.Bill
			items []db.BillItem
		)
		if id == 0 {
			bill = db.Bill{User: uid}
		} else {
			bill, items = db.GetBill(id, so)
		}
		res = map[string]interface{}{"bill": bill, "items": items}
		users := db.ListUsers(1, "id", "name")
		for _, u := range users {
			if u.ID == bill.User {
				res["user"] = u.Name
				break
			}
		}
		res["unused"] = db.GetUnused(id)
		jsonReply(w, res)
	case "POST":
		assert(r.ParseForm())
		//若status非空，修改进货单状态，目标状态根据当前状态确定，status的值无关紧要
		if r.FormValue("status") != "" {
			//TODO...
			db.SetInventoryByBill(id) //标记为完成并修改库存数量
			return
		}
		gid, _ := strconv.Atoi(r.FormValue("gid"))
		if gid > 0 { //提交单条目编辑
			items := db.GetBillItems(id, gid)
			item := items[0]
			cost, err := float(r.FormValue("cost"))
			if err == nil {
				item.Cost = cost
			}
			req, err := float(r.FormValue("request"))
			if err == nil {
				item.Request = req
			}
			cfm, err := float(r.FormValue("confirm"))
			if err == nil {
				item.Confirm = cfm
			}
			db.SetBillItem(item, 1)
			return
		}
	case "DELETE":
		ids := strings.SplitN(strings.TrimSpace(r.URL.Path[11:]), "/", 2)
		bid, _ := strconv.Atoi(ids[0])
		if bid <= 0 {
			panic(errors.New("invalid bom id"))
		}
		gid := 0
		if len(ids) == 2 && len(ids[1]) > 0 {
			gid, _ = strconv.Atoi(ids[1])
			if gid <= 0 {
				panic(errors.New("invalid goods id"))
			}
		}
		if gid == 0 {
			db.DeleteBill(bid)
		} else {
			db.DeleteBillItem(bid, gid)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
