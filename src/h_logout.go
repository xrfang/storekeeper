package main

import "net/http"

func logout(w http.ResponseWriter, r *http.Request) {
	T.SignOut(getCookie(r, "token"))
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
