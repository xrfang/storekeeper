package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xrfang/pindex"
)

type Goods struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Pinyin string  `json:"pinyin"`
	Stock  int     `json:"stock"`
	Cost   float64 `json:"cost"`
}

type skuQR struct {
	ID   int      `json:"id"`
	Name []string `json:"name"`
}

func (sqr skuQR) Caption() string {
	t := ""
	for _, s := range sqr.Name {
		if s[0] == '*' {
			t += s[1:]
		} else {
			t += s
		}
	}
	return t
}

func (sqr skuQR) Score() int {
	cnt := 1
	for _, s := range sqr.Name {
		if s[0] == '*' {
			cnt += len([]rune(s)) - 1
		}
	}
	return cnt
}

type SkuQueryResult struct {
	Found   []skuQR `json:"found"`
	Match   []skuQR `json:"match"`
	Missing []skuQR `json:"missing"`
}

func CountSKU() int {
	var cnt int
	assert(db.Get(&cnt, `SELECT COUNT(id) FROM goods`))
	return cnt
}

func whereAs(s string) (string, []interface{}) {
	cond := []string{`name=?`}
	args := []interface{}{s}
	rs := []rune(s)
	L := len(rs)
	for k := L; k > 0; k-- {
		i := 0
		for {
			cond = append(cond, `name LIKE ?`)
			args = append(args, "%"+string(rs[i:k+i])+"%")
			i++
			if i+k > L {
				break
			}
		}
	}
	return strings.Join(cond, " OR "), args
}

func markMatch(subj string, term string) []string {
	m := make(map[rune]bool)
	for _, r := range []rune(term) {
		m[r] = true
	}
	var mark, res []string
	for _, s := range []rune(subj) {
		if m[s] {
			mark = append(mark, "*"+string(s))
		} else {
			mark = append(mark, string(s))
		}
	}
	var t string
	for _, m := range mark {
		if t == "" {
			t = m
			continue
		}
		if t[0] == '*' {
			if m[0] == '*' {
				t += m[1:]
			} else {
				res = append(res, t)
				t = m
			}
		} else {
			if m[0] != '*' {
				t += m
			} else {
				res = append(res, t)
				t = m
			}
		}
	}
	if len(t) > 0 {
		res = append(res, t)
	}
	if len(res) == 1 && res[0][0] != '*' {
		return nil
	}
	return res
}

func pinInit(name string) string {
	var segs []string
	for _, p := range pindex.Encode(strings.ToLower(name)) {
		t := ""
		for _, x := range p {
			if x >= 'A' && x <= 'Z' {
				t += string(x)
			}
		}
		segs = append(segs, t)
	}
	return strings.Join(segs, " ")
}

func QuerySKU(terms []string) *SkuQueryResult {
	var goods []Goods
	if len(terms) == 0 {
		assert(db.Select(&goods, `SELECT id,name FROM goods ORDER BY pinyin`))
		var items []skuQR
		for _, g := range goods {
			items = append(items, skuQR{ID: g.ID, Name: []string{g.Name}})
		}
		return &SkuQueryResult{Found: items}
	}
	pm := make(map[string]skuQR)
	var qr SkuQueryResult
	addMatch := func(ms []skuQR) {
		for _, m := range ms {
			cap := m.Caption()
			p := pm[cap]
			if m.Score() > p.Score() {
				pm[cap] = m
			}
		}
	}
	for _, t := range terms {
		cond, args := whereAs(t)
		qry := fmt.Sprintf(`SELECT id,name,pinyin FROM goods WHERE %s`, cond)
		assert(db.Select(&goods, qry, args...))
		if len(goods) == 0 {
			qr.Missing = append(qr.Missing, skuQR{ID: 0, Name: []string{t}})
			continue
		}
		match := []skuQR{}
		for _, g := range goods {
			if g.Name == t {
				qr.Found = append(qr.Found, skuQR{ID: g.ID, Name: []string{g.Name}})
				match = nil
				break
			}
			mm := markMatch(g.Name, t)
			if mm != nil {
				match = append(match, skuQR{ID: g.ID, Name: mm})
			}
		}
		if match != nil {
			addMatch(match)
			qr.Missing = append(qr.Missing, skuQR{ID: 0, Name: []string{t}})
		}
	}
	for cap, m := range pm {
		found := false
		for _, f := range qr.Found {
			if cap == f.Caption() {
				found = true
				break
			}
		}
		if !found {
			qr.Match = append(qr.Match, m)
		}
	}
	return &qr
}

func UpdateSKUs(skus []Goods) {
	if len(skus) == 0 {
		return
	}
	tx := db.MustBegin()
	defer tx.Commit()
	for _, h := range skus {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			continue
		}
		var stmt string
		var args []interface{}
		if h.ID == 0 {
			h.Pinyin = pinInit(h.Name)
			stmt = `INSERT INTO goods (name,pinyin) VALUES (?,?)`
			args = []interface{}{h.Name, h.Pinyin}
		} else {
			h.Pinyin = strings.ToUpper(strings.TrimSpace(h.Pinyin))
			if strings.TrimSpace(h.Pinyin) == "" {
				h.Pinyin = pinInit(h.Name)
			}
			stmt = `UPDATE goods SET name=?,pinyin=?`
			args = []interface{}{h.Name, h.Pinyin}
			stmt += ` WHERE id=?`
			args = append(args, h.ID)
		}
		tx.MustExec(stmt, args...)
	}
	return
}

func GetSKUs(ids ...interface{}) (goods []Goods) {
	if len(ids) == 0 {
		assert(db.Select(&goods, `SELECT * FROM goods`))
	} else {
		assert(db.Select(&goods, `SELECT * FROM goods WHERE id IN
		    (?`+strings.Repeat(`,?`, len(ids)-1)+`)`, ids...))
	}
	return
}

func SearchGoods(term string) (goods []Goods) {
	name := strings.ToUpper(strings.TrimSpace(term))
	term = "%" + name + "%"
	args := []interface{}{term, term}
	qry := `SELECT * FROM goods WHERE name LIKE ? OR pinyin LIKE ?`
	assert(db.Select(&goods, qry, args...))
	idx := -1
	for i, g := range goods {
		ns := strings.FieldsFunc(g.Name, func(c rune) bool {
			return c == ' ' || c == '　' || c == '\t' || c == ',' || c == '，' ||
				c == '/' || c == '(' || c == ')' || c == '（' || c == '）'
		})
		for _, n := range ns {
			if strings.TrimSpace(n) == name {
				idx = i
				goods[i].Name = n
				break
			}
		}
	}
	if idx >= 0 {
		goods = []Goods{goods[idx]}
	}
	return
}

func FindSKU(idx string) (gs []Goods) {
	split := func(r rune) bool {
		return r < 'A' || r > 'Z'
	}
	pys := strings.FieldsFunc(strings.ToUpper(idx), split)
	if len(pys) == 0 {
		return
	}
	var cond []string
	var args []interface{}
	for _, py := range pys {
		cond = append(cond, `(pinyin LIKE ?)`)
		args = append(args, "%"+py+"%")
	}
	qry := `SELECT * FROM goods WHERE ` + strings.Join(cond, " OR ")
	assert(db.Select(&gs, qry, args...))
	return
}

func AnalyzeGoodsUsage() {
	usage := make(map[string][]map[string]interface{})
	qry := `SELECT b.updated,bi.gname,ABS(bi.request*b.sets) AS amount
		FROM bom_item bi,bom b WHERE bi.bom_id=b.id AND b.status>0 AND
		b.type=2 ORDER BY b.updated`
	rows, err := db.Queryx(qry)
	assert(err)
	defer rows.Close()
	for rows.Next() {
		r := make(map[string]interface{})
		assert(rows.MapScan(r))
		u := r["updated"].(time.Time)
		gu := usage[r["gname"].(string)]
		gu = append(gu, map[string]interface{}{
			"amount": r["amount"],
			"date":   u.Format("2006-01-02"),
		})
		usage[r["gname"].(string)] = gu
	}
	assert(rows.Err())
	je := json.NewEncoder(os.Stdout)
	je.SetIndent("", "    ")
	je.Encode(usage)
}
