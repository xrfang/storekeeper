package main

import (
	"net/http"
	"path"
)

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join(cf.WebRoot, "/imgs/favicon.png"))
}
