package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"storekeeper/db"

	"github.com/mdp/qrterminal"
	res "github.com/xrfang/go-res"
)

var L *StdLogger

func main() {
	conf := flag.String("conf", "", "configuration file")
	ver := flag.Bool("version", false, "show version info")
	init := flag.Bool("init", false, "initialize configuration")
	flag.Usage = func() {
		fmt.Printf("StoreKeeper %s\n", verinfo())
		fmt.Printf("\nUSAGE: %s OPTIONS\n", filepath.Base(os.Args[0]))
		fmt.Println("\nOPTIONS")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *ver {
		fmt.Println(verinfo())
		return
	}
	loadConfig(*conf)
	db.Initialize(cf.DBFile)
	if *init {
		key, err := otpGenKey("admin")
		assert(err)
		qrterminal.Generate(key.String(), qrterminal.L, os.Stdout)
		db.UpdateOTPKey("admin", key.Secret())
		return
	}
	L = NewLogger(cf.LogPath, 1024*1024, 10)
	L.SetDebug(cf.DbgMode)
	startBackup()
	if cf.OffDuty != "" {
		db.ReokeTokens()
	}
	policy := res.Verbatim
	if cf.DbgMode {
		policy = res.OverwriteIfNewer
	}
	assert(res.Extract(cf.WebRoot, policy))
	setupRoutes()
	svr := http.Server{
		Addr:         ":" + cf.Port,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	if cf.TLSCert != "" && cf.TLSPKey != "" {
		assert(svr.ListenAndServeTLS(cf.TLSCert, cf.TLSPKey))
	} else {
		assert(svr.ListenAndServe())
	}
}
