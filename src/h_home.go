package main

import (
	"net/http"
	"path"
	"storekeeper/db"
)

func home(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			httpError(w, e)
		}
	}()
	if r.URL.Path != "/" {
		http.ServeFile(w, r, path.Join(cf.WebRoot, r.URL.Path))
		return
	}
	uri := "/chkout"
	uid := db.CheckToken(getCookie(r, "token"))
	if uid == 0 {
		uri = "/login"
	}
	http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
}
