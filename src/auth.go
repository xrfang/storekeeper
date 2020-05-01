package main

import (
	"math/rand"
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

type token struct {
	t   string
	id  int
	upd time.Time
}

func (t *token) Expired() bool {
	return time.Since(t.upd).Seconds() > 1800
}

type tokenStore struct {
	store map[string]*token
	sync.Mutex
}

func (ts *tokenStore) Init() {
	ts.Lock()
	defer ts.Unlock()
	ts.store = make(map[string]*token)
	go func() {
		for {
			time.Sleep(time.Minute)
			ts.Lock()
			for s, t := range ts.store {
				if t.Expired() {
					delete(ts.store, s)
				}
			}
			ts.Unlock()
		}
	}()
}

func (ts *tokenStore) SignIn(id int) string {
	ts.Lock()
	defer ts.Unlock()
	tok := uuid(16)
	ts.store[tok] = &token{t: tok, id: id, upd: time.Now()}
	return tok
}

func (ts *tokenStore) SignOut(token string) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.store, token)
}

func (ts *tokenStore) Validate(token string) (ok bool, id int) {
	ts.Lock()
	defer ts.Unlock()
	t := ts.store[token]
	if t == nil {
		return false, 0
	}
	t.upd = time.Now()
	return true, t.id
}

var T tokenStore

func init() {
	T.Init()
}
