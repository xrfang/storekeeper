package main

import (
	"fmt"
	"net/http"

	"storekeeper/db"
)

func sku(w http.ResponseWriter, r *http.Request) {
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
	switch r.Method {
	case "GET":
		cnt := db.CountSKU()
		renderTemplate(w, "sku.html", struct{ Total int }{cnt})
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
