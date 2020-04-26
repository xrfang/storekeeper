package main

import (
	"os"
	"path"
	"path/filepath"
)

type Configuration struct {
	LogPath string
	DbgMode bool
	Port    string
	WebRoot string
	//TODO: define configuration items
	cfgFile string
	cfgPath string
}

func (c Configuration) abs(fn string) string {
	if fn == "" || path.IsAbs(fn) {
		return fn
	}
	p, _ := filepath.Abs(path.Join(c.cfgPath, fn))
	return p
}

func (c *Configuration) Load(fn string) {
	f, err := os.Open(fn)
	assert(err)
	defer f.Close()
	//TODO: load configuration from f
	c.cfgFile = c.abs(fn)
	c.cfgPath = path.Dir(c.cfgFile)
}

var cf Configuration

func loadConfig(cfgFile string) {
	cf.Port = "4372"
	cf.WebRoot = "../webroot"
	cf.LogPath = "../log"
	cf.Load(cfgFile)
	cf.WebRoot = cf.abs(cf.WebRoot)
	cf.LogPath = cf.abs(cf.LogPath)
}
