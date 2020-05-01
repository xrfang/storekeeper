package main

import (
	"fmt"
	"net/http"
	"strconv"

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
			Name:   r.FormValue("name"),
			Login:  r.FormValue("login"),
			Client: cli,
			Memo:   r.FormValue("memo"),
		}
		fmt.Fprintf(w, "%+v\n", u)
	case "DELETE":
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

/*
{
"account": [
"0"
],
"created": [
"正在创建"
],
"id": [
"0"
],
"login": [
"SomeOne"
],
"memo": [
"About that person"
],
"name": [
"路人甲"
]
}
*/
