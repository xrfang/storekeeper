package main

import (
	"net/http"
	"storekeeper/db"
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
	inuse, unuse := db.AnalyzeGoodsUsage()
	jsonReply(w, map[string]interface{}{
		"suggest": inuse,
		"against": unuse,
	})
}
