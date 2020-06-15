package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
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
			fmt.Fprintln(os.Stderr, "tokenStore.load:", e)
		}
	}()
	f, err := os.Open(ts.persist)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return
	}
	defer f.Close()
	lines := bufio.NewScanner(f)
	for lines.Scan() {
		kv := strings.Split(strings.TrimSpace(lines.Text()), "=")
		if len(kv) != 2 {
			continue
		}
		id, _ := strconv.Atoi(kv[1])
		fmt.Printf("id=%v\n", id)
		if id > 0 && len(kv[0]) > 0 {
			ts.store[kv[0]] = &token{t: kv[0], id: id, upd: time.Now()}
		}
	}
}

func (ts *tokenStore) save() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintln(os.Stderr, "tokenStore.save:", e)
		}
	}()
	f, err := os.Create(ts.persist)
	assert(err)
	defer func() { assert(f.Close()) }()
	for k, v := range ts.store {
		fmt.Fprintf(f, "%s=%d\n", k, v.id)
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
			for s, t := range ts.store {
				if t.Expired() {
					delete(ts.store, s)
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

func init() {
	T.Init()
}
