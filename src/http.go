package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path"
	"path/filepath"
	"time"
)

func getCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

func setCookie(w http.ResponseWriter, name, value string, age int) {
	exp := time.Now().Add(time.Duration(age) * time.Second)
	http.SetCookie(w, &http.Cookie{
		Name:    name,
		Value:   value,
		Path:    "/",
		MaxAge:  age,
		Expires: exp,
		Secure:  false,
	})
}

func renderTemplate(w http.ResponseWriter, tpl string, args interface{}) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	helper := template.FuncMap{
		"ver": func() string {
			return "V" + _G_REVS + "." + _G_HASH
		},
		"org": func() string {
			return cf.OrgName
		},
	}
	tDir := path.Join(cf.WebRoot, "templates")
	t, err := template.New("body").Funcs(helper).ParseFiles(path.Join(tDir, tpl))
	assert(err)
	sfs, err := filepath.Glob(path.Join(tDir, "shared/*"))
	if len(sfs) > 0 {
		t, err = t.ParseFiles(sfs...)
		assert(err)
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	assert(t.Execute(w, args))
}

func jsonReply(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	assert(json.NewEncoder(w).Encode(data))
}

func httpError(w http.ResponseWriter, e interface{}) {
	err := L.Err("%v", e)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
