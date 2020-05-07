package db

import (
	"fmt"
	"strings"
	"time"
)

/*
入库单状态：1=待提交；2=待收货；3=已入库
出库单状态：1=未配齐；2=未发货；3=未结账；4=已完成
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点
	User    int       `json:"user" db:"user_id"`
	Amount  float64   `json:"amount"`
	Markup  string    `json:"markup"`
	Fee     float64   `json:"fee"`
	Count   int       `json:"count"`
	Memo    string    `json:"memo"`
	Status  byte      `json:"status"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type BillItem struct {
	ID        int       `json:"id"`
	BomID     int       `json:"bid" db:"bom_id"`
	GoodsID   int       `json:"gid" db:"gid"`
	GoodsName string    `json:"gname" db:"gname"`
	Unit      string    `json:"unit"`
	Price     float64   `json:"price"`
	Count     int       `json:"count"`
	Status    int       `json:"status"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

//tpl模板可以指定的参数：ID、Type、User、Status
func ListBills(tpl *Bill) (bills []Bill, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	var cond []string
	var args []interface{}
	qry := `SELECT * FROM bom`
	if tpl == nil {
		assert(db.Select(&bills, qry))
		goto items
	}
	if tpl.ID > 0 {
		assert(db.Select(&bills, qry+` WHERE id=?`, tpl.ID))
		goto items
	}
	if tpl.Type > 0 {
		cond = append(cond, `type=?`)
		args = append(args, tpl.Type)
	}
	if tpl.User > 0 {
		cond = append(cond, `user_id=?`)
		args = append(args, tpl.User)
	}
	if tpl.Status > 0 {
		cond = append(cond, `status=?`)
		args = append(args, tpl.Status)
	}
	if len(cond) == 0 {
		qry += ` ORDER BY updated`
		assert(db.Select(&bills, qry))
	}
	qry += ` WHERE ` + strings.Join(cond, ` AND `)
	qry += ` ORDER BY updated`
	assert(db.Select(&bills, qry, args...))
items:
	if len(bills) == 0 {
		return
	}
	var ids []interface{}
	for _, b := range bills {
		ids = append(ids, b.ID)
	}
	qry = `SELECT bom_id,COUNT(id) FROM bom_item WHERE bom_id IN (?` +
		strings.Repeat(`,?`, len(ids)-1) + `) GROUP BY bom_id`
	rows, err := db.Query(qry, ids...)
	assert(err)
	defer rows.Close()
	cm := make(map[int]int)
	for rows.Next() {
		var bid, cnt int
		assert(rows.Scan(&bid, &cnt))
		cm[bid] = cnt
	}
	for i, b := range bills {
		bills[i].Count = cm[b.ID]
	}
	return
}

func GetBill(id int) (bill *Bill, items []BillItem, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return
}

type BomItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Pinyin string `json:"pinyin"`
	Unit   string `json:"unit"`
	Count  int    `json:"count"`
}

func SearchGoods(term string, bomType int) (res map[string][]BomItem, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	var terms []string
	for _, t := range strings.Split(term, " ") {
		t = strings.TrimSpace(t)
		if len(t) > 0 {
			terms = append(terms, t)
		}
	}
	if len(terms) == 0 {
		return nil, fmt.Errorf("SearchGoods: term is empty")
	}
	res = make(map[string][]BomItem)
	var cond []string
	var args []interface{}
	for _, t := range terms {
		t = "%" + t + "%"
		cond = append(cond, "name LIKE ?", "pinyin LIKE ?")
		args = append(args, t, t)
	}
	qry := `SELECT id,name,pinyin FROM goods WHERE ` + strings.Join(cond, " OR ")
	var bis []BomItem
	assert(db.Select(&bis, qry, args...))
	if bomType == 1 { //进货单，需要获取默认采购量
		args = []interface{}{bomType}
		for _, b := range bis {
			args = append(args, b.ID)
		}
		type hist struct {
			ID    int
			GID   int
			Unit  string
			Count int
		}
		var hs []hist
		assert(db.Select(&hs, `SELECT bom_item.id,gid,unit,count FROM bom_item JOIN bom ON
			bom_item.bom_id=bom.id WHERE bom.type=? AND gid IN (?`+strings.Repeat(`,?`,
			len(args)-2)+`) GROUP BY gid HAVING MAX(bom_item.id) ORDER BY bom_item.id`, args...))
		for i, b := range bis {
			for _, h := range hs {
				if h.GID == b.ID {
					bis[i].Unit = h.Unit
					bis[i].Count = h.Count
				}
			}
		}
	}
	for _, b := range bis {
		for _, t := range terms {
			if strings.Contains(b.Name, t) || strings.Contains(b.Pinyin, t) {
				res[t] = append(res[t], b)
			}
		}
	}
	for _, t := range terms {
		if _, ok := res[t]; !ok {
			res[t] = []BomItem{}
		}
	}
	return
}
