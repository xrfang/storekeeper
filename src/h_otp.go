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
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA256,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func otpShow(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			http.Error(w, e.(error).Error(), http.StatusInternalServerError)
		}
	}()
	ok, _ := T.Validate(getCookie(r, "token"))
	if !ok {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	id, _ := strconv.Atoi(r.URL.Path[5:])
	u, err := db.GetUser(id)
	assert(err)
	key, err := otpGenKey(u.Login)
	assert(err)
	qrCode, _ := qr.Encode(key.String(), qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 200, 200)
	var buf bytes.Buffer
	fmt.Fprint(&buf, "data:image/png;base64,")
	be := base64.NewEncoder(base64.StdEncoding, &buf)
	png.Encode(be, qrCode)
	be.Close()
	renderTemplate(w, "otp.html", struct{ 
		Name string
		Code string
	}{u.Name, buf.String()})
}
