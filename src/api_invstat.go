package main

import (
	"fmt"
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
	uid = 1 //DEBUG ONLY
	if uid == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	inuse, unuse := db.AnalyzeGoodsUsage()
	fmt.Fprintln(w, "建议采购：")
	for _, g := range inuse {
		fmt.Fprintf(w, "%s %v克 ", g.Name, g.Amount)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "===")
	fmt.Fprintln(w, "三个月未使用的药材及其当前库存:")
	for _, g := range unuse {
		fmt.Fprintf(w, "%s %v克 ", g.Name, g.Amount)
	}
	fmt.Fprintln(w)
}
