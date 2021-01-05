package main

import (
	"fmt"
	"net/http"
	"storekeeper/db"
	"strconv"
)

func finChkList(w http.ResponseWriter, r *http.Request) {
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
	renderTemplate(w, "finchk.html", nil)
}

func finChkEdit(w http.ResponseWriter, r *http.Request) {
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
	if db.InventoryWIP() != 0 {
		http.Error(w, "当前有未结束的盘点", http.StatusConflict)
		return
	}
	//_probe参数表示客户端使用jquery测试本页面是否可以跳转，如果盘点进行中则不允许访问
	if r.URL.Query().Get("_probe") != "" {
		fmt.Fprintln(w, "OK") //可以继续访问
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[8:])
	_, err := db.LedgerGet(id)
	assert(err)
	renderTemplate(w, "finchked.html", map[string]interface{}{"lid": id})
}
