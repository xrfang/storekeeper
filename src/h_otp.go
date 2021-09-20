package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"strconv"

	"storekeeper/db"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func otpGenKey(name string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      cf.OrgName,
		AccountName: name,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func otpShow(w http.ResponseWriter, r *http.Request) {
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
	id, _ := strconv.Atoi(r.URL.Path[5:])
	u := db.GetUser(id)
	if u.Client != 0 {
		http.Redirect(w, r, "/users", http.StatusTemporaryRedirect)
		return
	}
	key, err := otpGenKey(u.Login)
	assert(err)
	db.UpdateOTPKey(u.Login, key.Secret())
	qrCode, _ := qr.Encode(key.String(), qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 200, 200)
	var buf bytes.Buffer
	fmt.Fprint(&buf, "data:image/png;base64,")
	be := base64.NewEncoder(base64.StdEncoding, &buf)
	png.Encode(be, qrCode)
	be.Close()
	renderTemplate(w, "otp.html", struct {
		Name string
		Code string
	}{u.Name, buf.String()})
}
