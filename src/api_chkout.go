package main

import (
	"net/http"
	"storekeeper/db"
	"strconv"
	"time"
)

func apiChkOutList(w http.ResponseWriter, r *http.Request) {
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
	bu, _ := strconv.Atoi(r.URL.Query().Get("user"))
	um := make(map[int]string)
	for _, u := range db.ListUsers(1, "id", "name") {
		um[u.ID] = u.Name
	}
	thisMonth := time.Now().Format("2006-01")
	month := r.URL.Query().Get("month")
	if month == "" {
		month = thisMonth
	}
	summary := db.ListBillSummary(2, bu)
	if len(summary) == 0 || summary[0].Month != thisMonth {
		summary = append([]db.BillSummary{db.BillSummary{Month: thisMonth}}, summary...)
	}
	list := db.ListBills(2, bu, month)
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
