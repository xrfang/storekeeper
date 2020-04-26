package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	res "github.com/xrfang/go-res"
)

var L *StdLogger

func main() {
	conf := flag.String("conf", "", "configuration file")
	ver := flag.Bool("version", false, "show version info")
	init := flag.Bool("init", false, "initialize configuration")
	flag.Usage = func() {
		fmt.Printf("WebServer - Go WebServer Template %s\n", verinfo())
		fmt.Printf("\nUSAGE: %s OPTIONS\n", filepath.Base(os.Args[0]))
		fmt.Println("\nOPTIONS")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *ver {
		fmt.Println(verinfo())
		return
	}
	if *init {
		fmt.Println("TODO: initialize configuration")
		return
	}
	loadConfig(*conf)
	L = NewLogger(cf.LogPath, 1024*1024, 10)
	L.SetDebug(cf.DbgMode)
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
	assert(svr.ListenAndServe())
}
