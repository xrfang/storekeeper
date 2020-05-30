package main

import (
	"net/http"
)

func setupRoutes() {
	http.HandleFunc("/", home)
	http.HandleFunc("/api/set/", apiSetProp)
	http.HandleFunc("/api/get/", apiGetBill)
	http.HandleFunc("/api/chkin/", apiChkIn)
	http.HandleFunc("/api/sku", apiSkuList)
	http.HandleFunc("/api/sku/", apiSkuEdit)
	http.HandleFunc("/api/users", apiUsers)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/otp/", otpShow)
	http.HandleFunc("/chkin", chkInList)
	http.HandleFunc("/chkin/", chkInEdit)
	http.HandleFunc("/chkin/memo/", chkInSetMemo)
	http.HandleFunc("/chkout", chkOutList)
	http.HandleFunc("/chkout/", chkOutEdit)
	http.HandleFunc("/chkout/item/", chkOutEditItem)
	http.HandleFunc("/invchk", invChkList)
	http.HandleFunc("/invchk/", invChkEdit)
	http.HandleFunc("/invchk/item/", invChkEditItem)
	http.HandleFunc("/sku", sku)
	http.HandleFunc("/sku/", sku)
	http.HandleFunc("/users", users)
	http.HandleFunc("/users/", users)
}
