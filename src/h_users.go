package main

import (
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
	if u.ID == 1 {
		resp["mesg"] = "不可编辑管理员信息（但可以重置其登录密钥）"
		return resp
	}
	r := regexp.MustCompile(`^[0-9a-z]{5,32}$`)
	if !r.MatchString(u.Login) {
		resp["mesg"] = "登录标识必须由5~32个数字字母构成"
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
			httpError(w, e)
		}
	}()
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
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
			renderTemplate(w, "users.html", struct{ ID int }{uid})
			return
		}
		var ui UserInfo
		pu := db.GetPrimaryUsers()
		id, _ := strconv.Atoi(ids)
		if id > 0 {
			u := db.GetUser(id)
			ui = UserInfo{
				ID:      u.ID,
				Name:    u.Name,
				Login:   u.Login,
				Client:  u.Client,
				Memo:    u.Memo,
				Created: u.Created.Format("2006-01-02"),
			}
		}
		if ui.ID == 0 { //新增用户
			if uid == 1 {
				ui.AccList = pu
			} else {
				ui.Client = uid
				for _, p := range pu {
					if p.ID == uid {
						ui.AccList = []db.User{p}
						break
					}
				}
			}
		} else { //编辑现有用户
			for _, p := range pu {
				if p.ID == ui.Client {
					ui.AccList = []db.User{p}
					break
				}
			}
		}
		renderTemplate(w, "usered.html", ui)
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
			if db.UpdateUser(&u) {
				if strings.HasPrefix(resp["goto"].(string), "/otp/") {
					resp["goto"] = resp["goto"].(string) + strconv.Itoa(u.ID)
				}
			} else {
				resp["stat"] = false
				resp["mesg"] = "不能修改该用户的信息"
			}
		}
		jsonReply(w, resp)
	case "DELETE":
		var ids string
		if len(r.URL.Path) > 7 {
			ids = r.URL.Path[7:]
		}
		id, _ := strconv.Atoi(ids)
		if id <= 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if id == 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if uid != 1 {
			u := db.GetUser(id)
			if u.Client != uid {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		db.DeleteUser(id)
	}
}
