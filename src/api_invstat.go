package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func apiInvStat(w http.ResponseWriter, r *http.Request) {
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
	act, _ := strconv.Atoi(r.URL.Query().Get("act"))
	inuse, unuse := db.AnalyzeGoodsUsage()
	if act == 0 {
		jsonReply(w, map[string]interface{}{
			"suggest": inuse,
			"against": unuse,
		})
		return
	}
	id := db.SetBill(db.Bill{Type: 1, User: uid})
	var rx string
	for _, u := range inuse {
		rx += fmt.Sprintf("%v %v ", u.Name, u.Amount)
	}
	ps := db.GetPSItems(rx)
	for _, p := range ps {
		if len(p.Items) == 1 && p.Weight != nil && *p.Weight > 0 {
			db.SetBillItem(db.BillItem{
				BomID:     id,
				Cost:      p.Items[0].Cost,
				GoodsID:   p.Items[0].ID,
				GoodsName: p.Items[0].Name,
				Memo:      p.Memo,
				Request:   *p.Weight,
			}, 0)
		}
	}
	jsonReply(w, map[string]interface{}{"id": id})
}
