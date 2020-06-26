package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const LOG_FN = "log"

type StdLogger struct {
	dbgMode bool
	path    string
	size    int
	keep    int
	lines   []string
	sync.Mutex
}

type SysLog struct {
	Timestamp string   `json:"time"`
	Message   []string `json:"msgs"`
}

func NewLogger(path string, split, keep int) *StdLogger {
	if path != "" {
		path, err := filepath.Abs(path)
		assert(err)
		assert(os.MkdirAll(path, 0750))
	}
	sl := StdLogger{path: path, size: split, keep: keep}
	go func() {
		for {
			time.Sleep(time.Second)
			sl.flush()
			if sl.path != "" {
				sl.rotate()
			}
		}
	}()
	return &sl
}

func (sl *StdLogger) SetDebug(mode bool) {
	sl.dbgMode = mode
}

func (sl *StdLogger) fmt(format string, args ...interface{}) []string {
	ts := time.Now().Format("2006-01-02 15:04:05 ")
	pad := strings.Repeat(" ", len(ts))
	var msg []string
	for i, m := range strings.Split(fmt.Sprintf(format, args...), "\n") {
		m = strings.TrimRight(m, " \n\r\t")
		if m == "" {
			continue
		}
		if i == 0 {
			msg = append(msg, ts+m)
		} else {
			msg = append(msg, pad+m)
		}
	}
	return msg
}

func (sl *StdLogger) split(fn string) {
	defer func() {
		if e := recover(); e != nil {
			err := trace("[ERROR]StdLogger.split: %v", e)
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	st, err := os.Stat(fn)
	if err != nil || st.Size() < int64(sl.size) {
		return
	}
	dest := fn + "." + time.Now().Format("20060102-150405")
	assert(os.Rename(fn, dest))
	go func(fn string) {
		defer func() {
			if e := recover(); e != nil {
				err := trace("[ERROR]StdLogger.split.gzip: %v", e)
				fmt.Fprintln(os.Stderr, err)
				return
			}
			os.Remove(fn)
		}()
		f, err := os.Open(fn)
		assert(err)
		defer f.Close()
		g, err := os.Create(fn + ".gz")
		assert(err)
		defer g.Close()
		zw := gzip.NewWriter(g)
		defer func() {
			assert(zw.Close())
		}()
		_, err = io.Copy(zw, f)
		assert(err)
	}(dest)
}

func (sl *StdLogger) persist() {
	if sl.path == "" {
		return
	}
	fn := filepath.Join(sl.path, LOG_FN)
	sl.split(fn)
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR]StdLogger.persist:", err)
		return
	}
	defer f.Close()
	for _, c := range sl.lines {
		fmt.Fprintln(f, c)
	}
}

func (sl *StdLogger) flush() {
	sl.Lock()
	defer sl.Unlock()
	if len(sl.lines) > 0 {
		sl.persist()
		sl.lines = nil
	}
}

func (sl *StdLogger) rotate() {
	defer func() {
		if e := recover(); e != nil {
			err := trace("StdLogger.rotate: %v", e)
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	d, err := os.Open(sl.path)
	assert(err)
	defer d.Close()
	fis, err := d.Readdir(-1)
	assert(err)
	if len(fis) <= sl.keep {
		return
	}
	sort.Slice(fis, func(i, j int) bool {
		ti := fis[i].ModTime()
		tj := fis[j].ModTime()
		return ti.Before(tj)
	})
	for i := 0; i < len(fis)-sl.keep; i++ {
		fn := filepath.Join(sl.path, fis[i].Name())
		os.Remove(fn)
	}
}

func (sl *StdLogger) Dbg(format string, args ...interface{}) {
	if sl.dbgMode {
		sl.Log(format, args...)
	}
}

func (sl *StdLogger) Err(format string, args ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	var exp bool
	for _, a := range args {
		switch a.(type) {
		case error:
			exp = true
			format = "[EXCEPTION] " + format
		default:
			exp = false
		}
		if exp {
			break
		}
	}
	if !exp {
		format = "[ERROR] " + format
	}
	err := trace(format, args...)
	msg := sl.fmt(err.Error())
	for _, m := range msg {
		fmt.Fprintln(os.Stderr, m)
		sl.lines = append(sl.lines, m)
		if !sl.dbgMode && !exp {
			break
		}
	}
}

func (sl *StdLogger) Log(format string, args ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	msg := sl.fmt(format, args...)
	for _, m := range msg {
		fmt.Println(m)
	}
	sl.lines = append(sl.lines, msg...)
}
