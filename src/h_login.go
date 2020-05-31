package main

import (
	"net/http"

	"storekeeper/db"
)

func login(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	assert(r.ParseForm())
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	var mesg string

	if user != "" && pass != "" {
		id, err := db.CheckLogin(user, pass)
		if err == nil {
			setCookie(w, "token", T.SignIn(id), 86400*30) //有效期限30天
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		if err == db.ErrInvalidOTP {
			mesg = "用户名或密码错误"
		} else {
			L.Log("login: %v", err)
			mesg = "内部错误"
		}
	}
	if user == "" {
		user = getCookie(r, "user")
	}
	renderTemplate(w, "login.html", struct {
		User string
		Err  string
	}{user, mesg})
}
