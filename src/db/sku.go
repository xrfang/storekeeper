package db

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xrfang/pindex"
)

type Goods struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Pinyin string  `json:"pinyin"`
	Stock  float64 `json:"stock"`
	Cost   float64 `json:"cost"`
	Batch  float64 `json:"batch"`
	Rack   string  `json:"rack"`
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
		assert(db.Select(&goods, `SELECT * FROM goods ORDER BY pinyin`))
		var items []skuQR
		for _, g := range goods {
			name := g.Name
			if g.Rack != "" {
				name += "[" + g.Rack + "]"
			}
			items = append(items, skuQR{ID: g.ID, Name: []string{name}})
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
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	for _, h := range skus {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			continue
		}
		var stmt string
		var args []interface{}
		if h.ID == 0 {
			h.Pinyin = pinInit(h.Name)
			if h.Batch > 0 {
				stmt = `INSERT INTO goods (name,pinyin,batch) VALUES (?,?,?)`
				args = []interface{}{h.Name, h.Pinyin, h.Batch}
			} else {
				stmt = `INSERT INTO goods (name,pinyin) VALUES (?,?)`
				args = []interface{}{h.Name, h.Pinyin}
			}
		} else {
			h.Pinyin = strings.ToUpper(strings.TrimSpace(h.Pinyin))
			if strings.TrimSpace(h.Pinyin) == "" {
				h.Pinyin = pinInit(h.Name)
			}
			if h.Batch < 0 {
				h.Batch = 0
			}
			stmt = `UPDATE goods SET name=?,pinyin=?,batch=?,rack=?`
			args = []interface{}{h.Name, h.Pinyin, h.Batch, h.Rack}
			stmt += ` WHERE id=?`
			args = append(args, h.ID)
		}
		tx.MustExec(stmt, args...)
	}
	return
}

func UpdateRack(gid int, rack string) {
	res := db.MustExec(`UPDATE goods SET rack=? WHERE id=?`, rack, gid)
	ra, _ := res.RowsAffected()
	if ra <= 0 {
		panic(fmt.Errorf(`UpdateRack(%d, %s): no rows affected`, gid, rack))
	}
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

type UsageInfo struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Batch  int    `json:"batch"`
}

/*
AnalyzeGoodsUsage 分析库存需求。算法如下：

1. 计算需求等级分

首先，获取最近3个月的出库数据（例如今天为8月29日，起算日期就是5月29日）。按照药材与月份进行聚合。
以黄连为例，获得如下数据（用量、使用日期）：

	9 	2020-06-08
	21	2020-06-08
	70	2020-06-09
	15	2020-06-15
	20	2020-06-20
	15	2020-06-21
	84	2020-07-19
	15	2020-08-11
	84	2020-08-15
	42	2020-08-16
	70	2020-08-24

将这些用量加总再乘以使用次数，即获得黄连的“等级分”为4895分。基本逻辑是，药材使用次数越多，
未来被再次用到的概率越大。也就是说更看重使用的次数，其次才是每次的使用量。

2. 计算补货需求量

注意：补货数量算法假设平均一个月采购一次，每味药材的采购量为补货单位的倍数。

将药材3个月的最大使用量乘以月均使用次数，即得预测用量。例如黄连的预测用量为84*11/3=308克。
在计算最大使用量时有一个特殊逻辑：如果某药材的单剂使用量大于99克则认为是药材代购（例如某人
要求代购1000克白参），不计入最大使用量。

检查库存，如果小于预测用量则不用补货。否则，补充其差值。另外，在计算补货量的时候增加10克误
差量。意思是说，差值小于10克时不补货，在10～B+10克之间时采购B克，在B+10～2B+10克之间时采
购2B克，依次类推。例如，黄连的库存为80克，差值为228克，在10～509之间（采购单位B为500克），
输出建议采购量为500克。

3. 输出方式

本函数输出两个信息：

- 按照等级分由高到低排序的需要补货的药材及其建议补货量
- 三个月内没有任何出货的药材及其库存（滤除库存为0的药材）
*/
func AnalyzeGoodsUsage() ([]UsageInfo, []UsageInfo) {
	n := time.Now()
	since := time.Date(n.Year(), n.Month()-3, n.Day(), 0, 0, 0, 0, time.Local)
	usage := make(map[string][]map[string]int64)
	qry := `SELECT b.sets,bi.gname,ABS(bi.request*b.sets) AS amount
		FROM bom_item bi,bom b WHERE bi.bom_id=b.id AND b.status>0 
		AND	b.type=2 AND b.updated>=?`
	rows, err := db.Queryx(qry, since.Format("2006-01-02"))
	assert(err)
	active := make(map[string]bool)
	defer rows.Close()
	for rows.Next() {
		r := make(map[string]interface{})
		assert(rows.MapScan(r))
		gname := r["gname"].(string)
		active[gname] = true
		gu := usage[gname]
		gu = append(gu, map[string]int64{
			"amount": ival(r["amount"]),
			"sets":   ival(r["sets"]),
		})
		usage[gname] = gu
	}
	assert(rows.Err())
	used := make(map[string]UsageInfo)
	var inuse, unuse []UsageInfo
	var gs []Goods
	assert(db.Select(&gs, `SELECT * FROM goods`))
	for _, g := range gs {
		if active[g.Name] {
			used[g.Name] = UsageInfo{g.Name, int(g.Stock), int(g.Batch)}
		} else if g.Stock > 0 {
			unuse = append(unuse, UsageInfo{g.Name, int(g.Stock), int(g.Batch)})
		}
	}
	type survey struct {
		Score int
		Usage int
	}
	sm := make(map[string]survey)
	for g, u := range usage {
		if len(u) < 2 { //仅使用1次的药材不考虑采购
			continue
		}
		k := used[g] //药材的当前库存信息
		batch := int(k.Batch)
		if batch <= 0 { //该药材被设置为不建议采购
			continue
		}
		var total, max int
		for _, c := range u {
			amt := int(c["amount"])
			total += amt
			if amt > max {
				if amt/int(c["sets"]) < 100 {
					max = amt
				}
			}
		}
		s := survey{
			Score: total * len(u),
			Usage: max * len(u) / 3,
		}
		diff := s.Usage - int(k.Amount) - 9
		if diff > 0 {
			buy := diff / batch
			if diff%batch > 0 {
				buy++
			}
			inuse = append(inuse, UsageInfo{g, buy * batch, 1})
		}
		sm[g] = s
	}
	sort.Slice(inuse, func(i, j int) bool {
		si := sm[inuse[i].Name]
		sj := sm[inuse[j].Name]
		return si.Score > sj.Score
	})
	return inuse, unuse
}
