package main

import (
	"net/http"
	"storekeeper/db"
	"strconv"
)

func apiLedger(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			err := trace("%v", e)
			jsonReply(w, map[string]interface{}{
				"stat": false,
				"mesg": err.Error(),
			})
		}
	}()
	//TODO: add auth control
	//uid := db.CheckToken(getCookie(r, "token"))
	//if uid == 0 {
	//	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//	return
	//}
	switch r.Method {
	case "GET":
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		var data interface{}
		if id <= 0 {
			data = db.ListLedgers()
		} else {
			data = db.GetLedger(id)
		}
		jsonReply(w, map[string]interface{}{"stat": true, "data": data})
	case "POST":
		id := db.LedgerBills()
		jsonReply(w, map[string]interface{}{"stat": true, "data": id})
	case "DELETE":
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		db.UnledgerBills(id)
		jsonReply(w, map[string]interface{}{"stat": true})
	}
}
