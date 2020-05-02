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
	Client  int
	AccList []db.User
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
	redir := "/users"
	if u.ID == 0 {
		redir = "/otp/"
	}
	return map[string]interface{}{"stat": true, "goto": redir}
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
	switch r.Method {
	case "GET":
		var ids string
		if len(r.URL.Path) > 7 {
			ids = r.URL.Path[7:]
		}
		if ids == "" {
			renderTemplate(w, "users.html", struct{ID int}{uid})
		} else {
			id, _ := strconv.Atoi(ids)
			var ui UserInfo
			if id == 0 {
				if uid != 1 {
					ui.Client = uid
				}
			} else {
				u, err := db.GetUser(id)
				if err != nil {
					ui.Error = err.Error()
				} else {
					ui = UserInfo{
						ID:      u.ID,
						Name:    u.Name,
						Login:   u.Login,
						Client:  u.Client,
						Memo:    u.Memo,
						Created: u.Created.Format("2006-01-02"),
					}
					if uid == 1 {
						pu, err := db.GetPrimaryUsers()
						if err != nil {
							ui.Error = err.Error()
						} else {
							ui.AccList = pu
						}
					}
				}
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
			if uid != 1 {
				u.Client = -1 //禁止非admin用户设置u.Client
			}
			err := db.UpdateUser(&u)
			if err == nil {
				if strings.HasPrefix(resp["goto"].(string), "/otp/") {
					resp["goto"] = resp["goto"].(string) + strconv.Itoa(u.ID)
				}
			} else {
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
