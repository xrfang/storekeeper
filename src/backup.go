package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
)

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

func doBackup() {
	tag := time.Now().Format("2006-01-02_1504")
	bak := fmt.Sprintf("%s/%s.%s", cf.bkupDir, filepath.Base(cf.DBFile), tag)
	L.Dbg("doBackup => %s", bak)
	defer func() {
		if e := recover(); e != nil {
			L.Err("doBackup: %v", e)
		}
	}()
	cp(cf.DBFile, bak)
	fmt.Println("TODO: clean up old backup files and duplicates")
}

func startBackup() {
	c := cron.New()
	c.AddFunc("* * * * *", doBackup)
	c.Start()
}
