package main

import (
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/api/chkin/", apiChkIn)
	http.HandleFunc("/api/sku", apiSkuList)
	http.HandleFunc("/api/sku/", apiSkuEdit)
	http.HandleFunc("/api/users", apiUsers)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/otp/", otpShow)
	http.HandleFunc("/chkin", chkInList)
	http.HandleFunc("/chkin/", chkInEdit)
	http.HandleFunc("/chkout", chkOutList)
	http.HandleFunc("/chkout/", chkOutEdit)
	http.HandleFunc("/chkout/fee/", chkOutSetFee)
	http.HandleFunc("/chkout/mark/", chkOutSetMarkup)
	http.HandleFunc("/chkout/req/", chkOutSetRequester)
	http.HandleFunc("/chkout/item/", chkOutEditItem)
	http.HandleFunc("/inventory", inventory)
	http.HandleFunc("/sku", sku)
	http.HandleFunc("/sku/", sku)
	http.HandleFunc("/users", users)
	http.HandleFunc("/users/", users)
}
