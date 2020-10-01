package main

import (
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	LogPath string  `yaml:"log_path"`
	DbgMode bool    `yaml:"dbg_mode"`
	Markup  float64 `yaml:"markup"`
	Port    string  `yaml:"port"`
	WebRoot string  `yaml:"web_root"`
	DBFile  string  `yaml:"db_file"`
	OrgName string  `yaml:"org_name"`
	TLSCert string  `yaml:"tls_cert"`
	TLSPKey string  `yaml:"tls_pkey"`
	OffDuty string  `yaml:"off_duty"`
	cfgFile string
	cfgPath string
	bkupDir string
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
	cf.bkupDir = filepath.Join(cf.cfgPath, "backup")
	assert(os.MkdirAll(cf.bkupDir, 0755))
}

var cf Configuration

func loadConfig(cfgFile string) {
	cf.Port = "4372"
	cf.WebRoot = "../webroot"
	cf.LogPath = "../log"
	cf.DBFile = "herbs.db"
	cf.Markup = 30 //系统默认溢价率
	cf.OrgName = "Herb Store"
	cf.cfgPath = path.Dir(os.Args[0])
	if cfgFile != "" {
		cf.Load(cfgFile)
	}
	cf.WebRoot = cf.abs(cf.WebRoot)
	cf.LogPath = cf.abs(cf.LogPath)
	assert(os.MkdirAll(cf.LogPath, 0755))
	cf.DBFile = cf.abs(cf.DBFile)
	assert(os.MkdirAll(filepath.Dir(cf.DBFile), 0755))
	cf.TLSCert = cf.abs(cf.TLSCert)
	cf.TLSPKey = cf.abs(cf.TLSPKey)
}
