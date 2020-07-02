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
	TLSCert string `yaml:"tls_cert"`
	TLSPKey string `yaml:"tls_pkey"`
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
	assert(yd.Decode(c))
	fp, err := filepath.Abs(fn)
	assert(err)
	c.cfgPath = path.Dir(fp)
}

var cf Configuration

func loadConfig(cfgFile string) {
	cf.Port = "4372"
	cf.WebRoot = "../webroot"
	cf.LogPath = "../log"
	cf.DBFile = "herbs.db"
	cf.OrgName = "Herb Store"
	cf.cfgPath = path.Dir(os.Args[0])
	if cfgFile != "" {
		cf.Load(cfgFile)
	}
	cf.WebRoot = cf.abs(cf.WebRoot)
	cf.LogPath = cf.abs(cf.LogPath)
	cf.DBFile = cf.abs(cf.DBFile)
	cf.TLSCert = cf.abs(cf.TLSCert)
	cf.TLSPKey = cf.abs(cf.TLSPKey)
}
