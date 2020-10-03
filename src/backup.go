package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

func chksum(fn string) string {
	f, err := os.Open(fn)
	assert(err)
	defer f.Close()
	h := md5.New()
	_, err = io.Copy(h, f)
	assert(err)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func cp(src, dst string) {
	f, err := os.Open(src)
	assert(err)
	defer f.Close()
	g, err := os.Create(dst)
	assert(err)
	defer func() { assert(g.Close()) }()
	_, err = io.Copy(g, f)
	assert(err)
}

var (
	bkupreg map[string]bool
	mx      sync.Mutex
)

func bkupManifest() {
	bkupreg = make(map[string]bool)
	fns, _ := filepath.Glob(filepath.Join(cf.bkupDir, "*"))
	monthAgo := time.Now().Add(-30 * 24 * time.Hour)
	exp := 0
	for _, fn := range fns {
		sum := chksum(fn)
		s := strings.Split(fn, ".")
		t, err := time.Parse("2006-01-02_1504", s[len(s)-1])
		if err != nil || t.Before(monthAgo) {
			os.Remove(fn)
			exp++
			continue
		}
		bkupreg[sum] = true
	}
	L.Dbg("bkupManifest: registered=%d; expired=%d", len(bkupreg), exp)
}

func doBackup() {
	mx.Lock()
	defer func() {
		if e := recover(); e != nil {
			L.Err("doBackup: %v", e)
		}
		mx.Unlock()
	}()
	bkupManifest()
	cs := chksum(cf.DBFile)
	if bkupreg[cs] {
		L.Dbg("doBackup: DB not changed since last backup")
	} else {
		tag := time.Now().Format("2006-01-02_1504")
		bak := fmt.Sprintf("%s/%s.%s", cf.bkupDir, filepath.Base(cf.DBFile), tag)
		L.Dbg("doBackup => %s", bak)
		cp(cf.DBFile, bak)
	}
}

func startBackup() {
	c := cron.New()
	c.AddFunc("*/30 * * * *", doBackup)
	c.Start()
}
