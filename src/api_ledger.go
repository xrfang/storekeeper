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
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case "GET":
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		var data interface{}
		var err error
		if id <= 0 {
			data, err = db.LedgerList()
		} else {
			data, err = db.LedgerGet(id)
		}
		if err != nil {
			jsonReply(w, map[string]interface{}{"stat": false, "mesg": err.Error()})
		} else {
			jsonReply(w, map[string]interface{}{"stat": true, "data": data})
		}
	case "POST":
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if id > 0 {
			err := db.LedgerCls(id)
			if err == nil {
				jsonReply(w, map[string]interface{}{"stat": true})
			} else {
				jsonReply(w, map[string]interface{}{"stat": false, "mesg": err.Error()})
			}
			return
		}
		id, err := db.LedgerNew()
		if err == nil {
			jsonReply(w, map[string]interface{}{"stat": true, "data": id})
		} else {
			jsonReply(w, map[string]interface{}{"stat": false, "mesg": err.Error()})
		}
	case "DELETE":
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if err := db.LedgerDel(id); err != nil {
			jsonReply(w, map[string]interface{}{"stat": false, "mesg": err.Error()})
		} else {
			jsonReply(w, map[string]interface{}{"stat": true})
		}
	}
}
