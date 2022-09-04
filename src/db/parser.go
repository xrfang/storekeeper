package db

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type (
	item struct {
		ID   int     `json:"id"`
		Cost float64 `json:"cost"`
		Name string  `json:"name"`
		Rack string  `json:"rack"`
	}
	PSItem struct {
		Term   string   `json:"term"`
		Items  []item   `json:"items"`
		Weight *float64 `json:"weight"`
		Memo   string   `json:"memo"`
		Rack   string   `json:"rack"`
	}
	PSItems      []*PSItem
	Prescription struct {
		ID    int      `json:"id"`
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}
	normalizer struct {
		rm int //模式：0=直接替换；1=保留映射关系（可反向替换回来）
		rx *regexp.Regexp
		rv string
	}
)

// MatchItems 将Items与参数itm匹配，没有出现在itm中的条目会被删除
func (pi *PSItem) MatchItems(itm map[string]*BillItem) {
	var its []item
	for _, it := range pi.Items {
		if itm[it.Name] != nil {
			its = append(its, it)
		}
	}
	pi.Items = its
}

var norm = []normalizer{
	{0, regexp.MustCompile(`\s*，\s*`), ","},
	{0, regexp.MustCompile(`（\s*`), " ("},
	{0, regexp.MustCompile(`\s*）`), ") "},
	{0, regexp.MustCompile(`@`), ":"},
	{0, regexp.MustCompile(`\s*：\s*`), " :"},
	{1, regexp.MustCompile(`\(.*?\)`), ""},
	{1, regexp.MustCompile(`:\S+`), ""},
}

func fetchItems(term string) []item {
	var its []item
	assert(db.Select(&its, `SELECT id,cost,name,rack FROM goods WHERE
	    name LIKE ? OR pinyin LIKE ?`, `%`+term+`%`, `%`+term+`%`))
	for _, it := range its {
		if it.Name == term {
			return []item{it}
		}
	}
	return its
}

func GetPSItems(text string) PSItems {
	var subst []string
	cclass := func(r rune) int {
		if (r >= '0' && r <= '9') || r == '-' || r == '.' {
			return 1
		}
		return -1
	}
	for _, n := range norm {
		switch n.rm {
		case 0:
			text = n.rx.ReplaceAllString(text, n.rv)
		case 1:
			text = n.rx.ReplaceAllStringFunc(text, func(s string) string {
				no := len(subst)
				subst = append(subst, s)
				return " @" + string('A'+no) + " "
			})
		}
	}
	pc := 0 //前一字符种类：0=无前一字符；1=数字，-1=非数字
	var sb strings.Builder
	for _, r := range text {
		cc := cclass(r)
		if pc != cc && pc != 0 {
			sb.WriteString(" ")
		}
		sb.WriteRune(r)
		pc = cc
	}
	text = sb.String()
	ss := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	})
	var ps PSItems
	var p *PSItem
	for _, s := range ss {
		w, err := float(s)
		if err == nil {
			if p != nil {
				p.Weight = new(float64)
				*p.Weight = w
			}
		} else if s[0] == '@' {
			s = subst[s[1]-'A']
			if s[0] == '(' {
				if p != nil {
					p.Memo = s[1 : len(s)-1]
				}
			} else if s[0] == ':' {
				if p != nil {
					p.Rack = strings.ToUpper(s[1:])
				}
			}
		} else {
			if p != nil {
				ps = append(ps, p)
			}
			if s == "克" {
				p = nil
			} else {
				p = new(PSItem)
				p.Term = strings.ToUpper(s)
				p.Items = fetchItems(p.Term)
			}
		}
	}
	if p != nil {
		ps = append(ps, p)
	}
	return ps
}

func GetUnused(bid int) []UsageInfo {
	_, nu := AnalyzeGoodsUsage()
	stock := make(map[string]int)
	for _, u := range nu {
		stock[u.Name] = u.Amount
	}
	var unused []UsageInfo
	_, items := GetBill(bid, 0)
	for _, it := range items {
		amt := stock[it.GoodsName]
		if amt > 0 {
			unused = append(unused, UsageInfo{
				Name:   it.GoodsName,
				Amount: amt,
				Batch:  1,
			})
		}
	}
	return unused
}

func GetPrevRx(ps PSItems) []Prescription {
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
	var rxs []Prescription
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
		var rx Prescription
		rx.ID, _ = strconv.Atoi(bis[0].Seq)
		rx.Name = bis[0].Name
		for _, b := range bis[1:] {
			rx.Items = append(rx.Items, b.Name)
		}
		rxs = append(rxs, rx)
	}
	sort.Slice(rxs, func(i, j int) bool { return rxs[i].ID > rxs[j].ID })
	return rxs
}
