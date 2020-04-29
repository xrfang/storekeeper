package main

import (
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	LogPath string `yaml:"log_path"`
	DbgMode bool   `yaml:"dbg_mode"`
	Port    string `yaml:"port"`
	WebRoot string `yaml:"web_root"`
	DBFile  string `yaml:"db_file"`
	OrgName string `yaml:"org_name"`
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
	yd := yaml.NewDecoder(f)
	yd.Decode(c)
	c.cfgPath = path.Dir(c.abs(fn))
}

var cf Configuration

func loadConfig(cfgFile string) {
	cf.Port = "4372"
	cf.WebRoot = "../webroot"
	cf.LogPath = "../log"
	cf.DBFile = "herbs.db"
	cf.OrgName = "Herbal Store"
	cf.cfgPath = path.Dir(os.Args[0])
	if cfgFile != "" {
		cf.Load(cfgFile)
	}
	cf.WebRoot = cf.abs(cf.WebRoot)
	cf.LogPath = cf.abs(cf.LogPath)
	cf.DBFile = cf.abs(cf.DBFile)
}
