package db

import (
	"errors"
	"strings"
	"time"
)

var ErrItemAlreadyExists = errors.New("item already exists")

/*
入库单状态：1=待提交；2=待收货；3=已入库
出库单状态：1=未配齐；2=未发货；3=未结账；4=已完成
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点
	User    int       `json:"user" db:"user_id"`
	Charge  float64   `json:"charge"`
	Cost    string    `json:"cost"`
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

func GetBill(id int, itmOrd int) (bill Bill, items []BillItem, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = trace("%v", e)
		}
	}()
	assert(db.Get(&bill, `SELECT * FROM bom WHERE id=?`, id))
	if itmOrd == 0 {
		assert(db.Select(&items, `SELECT * FROM bom_item WHERE bom_id=? ORDER BY id DESC`, id))
	} else {
		assert(db.Select(&items, `SELECT bi.* FROM bom_item bi JOIN goods g ON g.id=gid
		    WHERE bom_id=? ORDER BY g.pinyin`, id))
	}
	return
}

func SearchGoods(term string) (goods []Goods, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	name := strings.ToUpper(strings.TrimSpace(term))
	term = "%" + name + "%"
	args := []interface{}{term, term}
	qry := `SELECT id,name FROM goods WHERE name LIKE ? OR pinyin LIKE ?`
	assert(db.Select(&goods, qry, args...))
	if len(goods) == 1 {
		ns := strings.FieldsFunc(goods[0].Name, func(c rune) bool {
			return c == ' ' || c == '　' || c == '\t' || c == ',' || c == '，' ||
				c == '/' || c == '(' || c == ')' || c == '（' || c == '）'
		})
		for _, n := range ns {
			if strings.TrimSpace(n) == name {
				goods[0].Name = n
				break
			}
		}
	}
	return
}

func AddGoodsToBill(b Bill, gid int, gname string, cnt int) (id int, err error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	if b.ID == 0 {
		res := tx.MustExec(`INSERT INTO bom (type,user_id,fee,memo,status) VALUES
			(?,?,?,?1)`, b.Type, b.User, b.Fee, b.Memo)
		rid, err := res.LastInsertId()
		assert(err)
		b.ID = int(rid)
	} else {
		tx.MustExec(`UPDATE bom SET fee=?, memo=? WHERE ID=?`, b.Fee, b.Memo, b.ID)
	}
	var bi int
	assert(tx.Get(&bi, `SELECT COUNT(id) FROM bom_item WHERE bom_id=? AND gid=?`, b.ID, gid))
	if bi > 0 {
		panic(ErrItemAlreadyExists)
	}
	tx.MustExec(`INSERT INTO bom_item (bom_id, gid, gname,count) VALUES
		(?,?,?,?)`, b.ID, gid, gname, cnt)
	return b.ID, nil
}
