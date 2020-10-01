package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

func doBackup() {
	L.Log("TODO: backup database")
	bak := fmt.Sprintf("%s/herbs.db.%s", cf.bkupDir, time.Now().Format("20060102150405"))
	fmt.Printf("copy %s => %s\n", cf.DBFile, bak)
}

func startBackup() {
	c := cron.New()
	c.AddFunc("* * * * *", doBackup)
	c.Start()
}
