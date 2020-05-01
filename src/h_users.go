package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"storekeeper/db"
)

type UserInfo struct {
	ID      int
	Name    string
	Login   string
	Memo    string
	AccID   int
	AccList map[int]string
	Created string
	Error   string
}

func chkUser(u *db.User) map[string]interface{} {
	resp := map[string]interface{}{"stat": false}
	if len(u.Name) > 32 {
		resp["mesg"] = "姓名过长"
		return resp
	}
	r := regexp.MustCompile(`^[0-9a-z]{6,32}$`)
	if !r.MatchString(u.Login) {
		resp["mesg"] = "登录标识必须由6~32个数字字母构成"
		return resp
	}
	return map[string]interface{}{"stat": true, "goto": "/users"}
}

func users(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, uid := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	_ = uid
	switch r.Method {
	case "GET":
		ids := r.URL.Path[6:]
		if ids == "" {
			renderTemplate(w, "users.html", nil)
		} else {
			err := r.URL.Query().Get("err")
			id, _ := strconv.Atoi(ids)
			ui := UserInfo{
				ID:    id,
				Name:  "路人甲",
				Login: "SomeOne",
				AccID: 0,
				AccList: map[int]string{
					2: "陈萌慧",
					3: "石劲敏",
				},
				Memo:    "About that person",
				Created: "正在创建",
				Error:   err,
			}
			renderTemplate(w, "usered.html", ui)
		}
	case "POST":
		assert(r.ParseForm())
		id, _ := strconv.Atoi(r.FormValue("id"))
		cli, _ := strconv.Atoi(r.FormValue("client"))
		u := db.User{
			ID:     id,
			Name:   strings.TrimSpace(r.FormValue("name")),
			Login:  strings.ToLower(strings.TrimSpace(r.FormValue("login"))),
			Client: cli,
			Memo:   r.FormValue("memo"),
		}
		resp := chkUser(&u)
		if resp["stat"].(bool) {
			//TODO: 禁止非admin用户设置u.Client
			err := db.UpdateUser(&u)
			if err != nil {
				resp["stat"] = false
				resp["mesg"] = err.Error()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case "DELETE":
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
