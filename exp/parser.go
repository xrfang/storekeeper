package main

import (
	"regexp"
	"strconv"
	"strings"
)

type item struct {
	ID   int
	Name string
}
type psitem struct {
	Term   string
	Items  []item
	Weight int
	Memo   string
}

var rs []*regexp.Regexp = []*regexp.Regexp{
	regexp.MustCompile(`\s+`),
	regexp.MustCompile(`（\s*`),
	regexp.MustCompile(`\s*）`),
}

func fetchItems(term string) []item {
	var its []item
	assert(db.Select(&its, `SELECT id,name FROM goods WHERE name=? OR
		pinyin LIKE ?`, term, `%`+term+`%`))
	return its
}

func GetPSItems(text string) []*psitem {
	text = strings.TrimSpace(text)
	text = rs[1].ReplaceAllString(text, " (")
	text = rs[2].ReplaceAllString(text, ") ")
	ss := rs[0].Split(text, -1)
	if len(ss) == 0 || len(ss) == 1 && ss[0] == "" {
		return nil
	}
	var ps []*psitem
	var p *psitem
	for _, s := range ss {
		w, err := strconv.Atoi(s)
		if err == nil {
			if p != nil && w > 0 {
				p.Weight = w
			}
		} else if s[0] == '(' && s[len(s)-1] == ')' {
			if p != nil {
				p.Memo = s[1 : len(s)-1]
			}
		} else {
			if p != nil {
				ps = append(ps, p)
			}
			p = new(psitem)
			p.Term = strings.ToUpper(s)
			p.Items = fetchItems(p.Term)
		}
	}
	if p != nil {
		ps = append(ps, p)
	}
	return ps
}
