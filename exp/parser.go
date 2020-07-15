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
type psitems []*psitem

type prescription struct {
	ID    int
	Name  string
	Items []string
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

func GetPSItems(text string) psitems {
	text = strings.TrimSpace(text)
	text = rs[1].ReplaceAllString(text, " (")
	text = rs[2].ReplaceAllString(text, ") ")
	ss := rs[0].Split(text, -1)
	if len(ss) == 0 || len(ss) == 1 && ss[0] == "" {
		return nil
	}
	var ps psitems
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

func GetPrevRx(ps psitems) []prescription {
	var ids []interface{}
	for _, p := range ps {
		if len(p.Items) == 0 {
			continue
		}
		var args []interface{}
		for _, it := range p.Items {
			args = append(args, it.ID)
		}
		qry := `SELECT DISTINCT bom_id FROM bom_item,bom WHERE bom.id=bom_item.bom_id
			AND bom.type=2 AND gid IN (?` + strings.Repeat(`,?`, len(p.Items)-1) + `)`
		if len(ids) > 0 {
			args = append(args, ids...)
			qry += ` AND bom.id in (?` + strings.Repeat(`,?`, len(ids)-1) + `)`
		}
		ids = []interface{}{}
		assert(db.Select(&ids, qry, args...))
		if len(ids) == 0 {
			return nil
		}
	}
	var rxs []prescription
	type binfo struct {
		Seq  string
		Name string
	}
	for _, id := range ids {
		var bis []binfo
		qry := `SELECT bom.id AS seq,name FROM user,bom WHERE user.id=bom.user_id AND bom.id=? 
			UNION SELECT pinyin,gname FROM bom_item,goods WHERE goods.id=bom_item.gid AND
			bom_id=? ORDER BY seq`
		assert(db.Select(&bis, qry, id, id))
		var rx prescription
		rx.ID, _ = strconv.Atoi(bis[0].Seq)
		rx.Name = bis[0].Name
		for _, b := range bis[1:] {
			rx.Items = append(rx.Items, b.Name)
		}
		rxs = append(rxs, rx)
	}
	return rxs
}
