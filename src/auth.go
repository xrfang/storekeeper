package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"storekeeper/db"
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

func (t *token) expired() bool {
	return time.Since(t.upd).Hours() > 168 //最长7天未使用则失效
}

type tokenStore struct {
	store   map[string]*token
	changed bool
	persist string
	sync.Mutex
}

func (ts *tokenStore) load() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintln(os.Stderr, trace("tokenStore.load: %v", e))
		}
	}()
	for _, u := range db.ListUsers(1, "id", "session") {
		kv := strings.SplitN(u.Session, ",", 2)
		if len(kv) < 2 {
			continue
		}
		up, _ := strconv.Atoi(kv[0])
		ut := time.Unix(int64(up), 0)
		ts.store[kv[1]] = &token{t: kv[1], id: u.ID, upd: ut}
	}
}

func (ts *tokenStore) save() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintln(os.Stderr, "tokenStore.save:", e)
		}
	}()
	for _, t := range ts.store {
		db.SetAccessToken(t.id, t.t, t.upd)
	}
}

func (ts *tokenStore) Init() {
	ts.Lock()
	defer ts.Unlock()
	ts.persist = path.Join(path.Dir(os.Args[0]), "tokens")
	ts.store = make(map[string]*token)
	ts.load()
	go func() {
		for {
			time.Sleep(time.Minute)
			ts.Lock()
			for _, t := range ts.store {
				if t.expired() {
					t.t = ""
					ts.changed = true
				}
			}
			if ts.changed {
				ts.save()
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
	ts.changed = true
	return tok
}

func (ts *tokenStore) SignOut(token string) {
	ts.Lock()
	defer ts.Unlock()
	delete(ts.store, token)
	ts.changed = true
}

func (ts *tokenStore) Validate(token string) (ok bool, id int) {
	ts.Lock()
	defer ts.Unlock()
	t := ts.store[token]
	if t == nil {
		return false, 0
	}
	t.upd = time.Now()
	ts.changed = true
	return true, t.id
}

var T tokenStore
