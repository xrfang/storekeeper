package db

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	ErrItemAlreadyExists = errors.New("item already exists")
	iop                  sync.Mutex
)

/*
入库单状态：0=未完成；1=已完成
出库单状态：0=未完成；1=已完成；2=已收款
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点
	User    int       `json:"user" db:"user_id"`
	Markup  int       `json:"markup"`
	Fee     float64   `json:"fee"`
	Sets    int       `json:"sets"`
	Cost    float64   `json:"cost"`  //非数据库条目，实时计算，表示单剂药的成本
	Count   int       `json:"count"` //非数据库条目，实时计算
	Memo    string    `json:"memo"`
	Status  int       `json:"status"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type BillItem struct {
	ID        int       `json:"id"`
	BomID     int       `json:"bid" db:"bom_id"`
	GoodsID   int       `json:"gid" db:"gid"`
	GoodsName string    `json:"gname" db:"gname"`
	Cost      float64   `json:"cost"`
	Request   int       `json:"request"`
	Confirm   int       `json:"confirm"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

func RemoveEmptyBills() {
	_, err := db.Exec(`DELETE FROM bom WHERE NOT id IN 
		(SELECT distinct bom_id FROM bom_item)`)
	assert(err)
}

//tpl模板可以指定的参数：ID、Type、User、Status
func ListBills(tpl *Bill) (bills []Bill) {
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
	if tpl.Status >= 0 {
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
	qry = `SELECT bom_id,COUNT(id),SUM(ABS(cost)*ABS(request)) FROM bom_item WHERE 
	    bom_id IN (?` + strings.Repeat(`,?`, len(ids)-1) + `) GROUP BY bom_id`
	rows, err := db.Query(qry, ids...)
	assert(err)
	defer rows.Close()
	type summary struct {
		cnt int
		sum float64
	}
	cm := make(map[int]summary)
	for rows.Next() {
		var bid int
		var s summary
		assert(rows.Scan(&bid, &s.cnt, &s.sum))
		cm[bid] = s
	}
	for i, b := range bills {
		bills[i].Count = cm[b.ID].cnt
		bills[i].Cost = cm[b.ID].sum
	}
	return
}

func GetBill(id int, itmOrd int) (bill Bill, items []BillItem) {
	assert(db.Get(&bill, `SELECT * FROM bom WHERE id=?`, id))
	switch itmOrd {
	case 0:
		assert(db.Select(&items, `SELECT * FROM bom_item WHERE bom_id=? ORDER BY id DESC`, id))
	case 1:
		assert(db.Select(&items, `SELECT bi.* FROM bom_item bi JOIN goods g ON g.id=gid
		    WHERE bom_id=? ORDER BY g.pinyin`, id))
	default: //除以上两种itmOrd外，不返回items
		return
	}
	bill.Count = len(items)
	for i, it := range items {
		if bill.Type == 2 {
			items[i].Request = -items[i].Request
			items[i].Confirm = -items[i].Confirm
		}
		bill.Cost += math.Abs(it.Cost * float64(it.Request))
	}
	return
}

func GetBillItems(bid int, gid ...interface{}) (items []BillItem) {
	if len(gid) == 0 {
		return nil
	}
	ids := append([]interface{}{bid}, gid...)
	assert(db.Select(&items, `SELECT * FROM bom_item WHERE bom_id=? AND 
	    gid IN (?`+strings.Repeat(`,?`, len(gid)-1)+`)`, ids...))
	return
}

func SetBill(b Bill) (id int) {
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	if b.ID == 0 {
		res := tx.MustExec(`INSERT INTO bom (type,user_id,markup,fee,memo,sets) VALUES
		    (?,?,?,?,?,?)`, b.Type, b.User, b.Markup, b.Fee, b.Memo, b.Sets)
		id, err := res.LastInsertId()
		assert(err)
		return int(id)
	}
	tx.MustExec(`UPDATE bom SET user_id=?,markup=?,fee=?,memo=?,sets=?,status=?
		WHERE ID=?`, b.User, b.Markup, b.Fee, b.Memo, b.Sets, b.Status, b.ID)
	if b.Status > 0 && b.Type == 1 { //TODO：确认入库（status>0）时需要修改库存
		var bis []BillItem
		assert(tx.Select(&bis, `SELECT gid,confirm FROM bom_item WHERE bom_id=?`, b.ID))
		for _, bi := range bis {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, bi.Confirm, bi.GoodsID)
		}
	}
	return b.ID
}

func SetBillItem(bi BillItem, mode int) bool {
	b, _ := GetBill(bi.BomID, -1)
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	if b.Type == 2 { //出库单，数量转化为负值
		if bi.Request > 0 {
			bi.Request = -bi.Request
		}
		if bi.Confirm > 0 {
			bi.Confirm = -bi.Confirm
		}
	}
	switch mode {
	case 0: //insert
		var cnt int
		assert(tx.Get(&cnt, `SELECT COUNT(id) FROM bom_item WHERE bom_id=? AND gid=?`, bi.BomID, bi.GoodsID))
		if cnt > 0 {
			return true //该条目原本已经存在
		}
	case 1: //update
		tx.MustExec(`DELETE FROM bom_item WHERE bom_id=? AND gid=?`, bi.BomID, bi.GoodsID)
	}
	tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,cost,request,confirm) VALUES (?,?,?,?,?,?)`,
		bi.BomID, bi.GoodsID, bi.GoodsName, bi.Cost, bi.Request, bi.Confirm)
	if b.Type == 1 { //入库单，用当前价格更新药品单价
		tx.MustExec(`UPDATE goods SET cost=? WHERE id=?`, bi.Cost, bi.GoodsID)
	}
	return false //该条目原本不存在或被更新
}

func DeleteBill(bid int) {
	_, err := db.Exec(`DELETE FROM bom WHERE id=?`, bid)
	assert(err)
}

func DeleteBillItem(bid, gid int) {
	_, err := db.Exec(`DELETE FROM bom_item WHERE bom_id=? AND gid=?`, bid, gid)
	assert(err)
}

func SetInventoryByBill(bid, stat int) {
	iop.Lock()
	defer iop.Unlock()
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	var b Bill
	assert(tx.Get(&b, `SELECT * FROM bom WHERE id=?`, bid))
	switch b.Status {
	case 0:
		if stat != 1 {
			panic(fmt.Errorf("bill#%d.status=0; stat=%d", bid, stat))
		}
	case 1:
		if stat != 2 {
			panic(fmt.Errorf("bill#%d.status=1; stat=%d", bid, stat))
		}
		if b.Type != 2 {
			panic(fmt.Errorf("cannot set stat=2 for bill type %d", b.Type))
		}
	default:
		panic(fmt.Errorf("bill#%d.status=%d; stat=%d", bid, b.Status, stat))
	}
	switch b.Type {
	case 2:
		tx.MustExec(`UPDATE bom SET status=? WHERE id=?`, stat, b.ID)
		if stat != 1 {
			return
		}
		type billReq struct {
			ID        int
			Stock     int
			Requested int
		}
		var br []billReq
		assert(tx.Select(&br, `SELECT g.id,g.stock,ABS(bi.request)*b.sets AS requested
			FROM goods g,bom_item bi,bom b WHERE b.id=bi.bom_id AND bi.gid=g.id AND
			b.id=?`, bid))
		for _, r := range br {
			if r.Requested <= r.Stock {
				tx.MustExec(`UPDATE goods SET stock=stock-? WHERE id=?`, r.Requested, r.ID)
				tx.MustExec(`UPDATE bom_item SET confirm=request WHERE gid=? AND bom_id=?`,
					r.ID, bid)
			} else {
				confirm := -r.Stock / b.Sets
				tx.MustExec(`UPDATE goods SET stock=0 WHERE id=?`, r.ID)
				tx.MustExec(`UPDATE bom_item SET confirm=? WHERE gid=? AND bom_id=?`,
					confirm, r.ID, bid)
			}
		}
	case 3:
		tx.MustExec(`UPDATE bom SET status=? WHERE id=?`, stat, b.ID)
		if stat != 1 {
			return
		}
		var bis []BillItem
		assert(tx.Select(&bis, `SELECT gid,confirm FROM bom_item WHERE bom_id=?`, bid))
		for _, bi := range bis {
			tx.MustExec(`UPDATE goods SET stock=? WHERE id=?`, bi.Confirm, bi.GoodsID)
		}
	default:
		panic(fmt.Errorf("unsupported bill type %v", b.Type))
	}
}

func UpdateInventory(bid int) {
	iop.Lock()
	defer iop.Unlock()
	bill, _ := GetBill(bid, -1)
	if bill.Status != 0 {
		return
	}
	var bis []BillItem
	bim := make(map[int]bool)
	assert(db.Select(&bis, `SELECT gid FROM bom_item WHERE bom_id=?`, bid))
	for _, bi := range bis {
		bim[bi.GoodsID] = true
	}
	var gs []Goods
	assert(db.Select(&gs, `SELECT id,name,stock,cost FROM goods`))
	tx := db.MustBegin()
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	for _, g := range gs {
		if bim[g.ID] {
			tx.MustExec(`UPDATE bom_item SET request=?,cost=? WHERE bom_id=? AND gid=?`,
				g.Stock, g.Cost, bid, g.ID)
		} else if g.Stock > 0 {
			tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,request,cost,confirm)
			    VALUES (?,?,?,?,?,?)`, bid, g.ID, g.Name, g.Stock, g.Cost, 0)
		}
	}
}
