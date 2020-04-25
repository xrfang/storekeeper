package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func uuid(L int) string {
	cs := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	buf := make([]byte, L)
	rand.Read(buf)
	for i := 0; i < L; i++ {
		buf[i] = cs[buf[i]%62]
	}
	return string(buf)
}

var T sync.Map

func genToken() string {
	now := time.Now()
	T.Range(func(key, val interface{}) bool {
		if time.Since(val.(time.Time)).Seconds() > 3600 {
			T.Delete(key) //token固定1小时有效
		}
		return true
	})
	tok := uuid(16)
	T.Store(tok, now)
	return tok
}

func validate(r *http.Request) bool {
	created, ok := T.Load(getCookie(r, "token"))
	if !ok {
		return false
	}
	return time.Since(created.(time.Time)).Seconds() < 3600
}
