package main

import (
	"net/http"

	"storekeeper/db"
)

func login(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	assert(r.ParseForm())
	user := r.Form.Get("user")
	pass := r.Form.Get("pass")
	var mesg string

	if user != "" && pass != "" {
		var tok string
		var err error
		if cf.OffDuty == "" {
			tok, err = db.Login(user, pass)
		} else {
			if pass == cf.OffDuty {
				tok, err = db.MaintenanceLogin(user)
			} else {
				err = db.ErrOutOfService
			}
		}
		if err == nil {
			setCookie(w, "token", tok, 86400*30) //有效期限30天
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		switch err {
		case db.ErrInvalidOTP:
			mesg = "用户名或密码错误"
		case db.ErrOutOfService:
			mesg = "系统维护中，请联系管理员"
		default:
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
