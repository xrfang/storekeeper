package db

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

/*
入库单状态：0=未完成；1=已完成
出库单状态：0=未完成；1=已完成；2=已收款
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点
	User    int       `json:"user" db:"user_id"`
	Markup  float64   `json:"markup"`
	Fee     float64   `json:"fee"`
	Sets    int       `json:"sets"`
	Cost    float64   `json:"cost"`  //非数据库条目，实时计算，表示单剂药的成本
	Count   int       `json:"count"` //非数据库条目，实时计算
	Memo    string    `json:"memo"`
	Status  int       `json:"status"`
	Courier string    `json:"courier"`
	Paid    float64   `json:"paid"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type BillSummary struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type BillItem struct {
	ID        int       `json:"id"`
	BomID     int       `json:"bid" db:"bom_id"`
	GoodsID   int       `json:"gid" db:"gid"`
	GoodsName string    `json:"gname" db:"gname"`
	Cost      float64   `json:"cost"`
	Request   float64   `json:"request"`
	Confirm   float64   `json:"confirm"`
	Flag      int       `json:"flag"`
	Memo      string    `json:"memo"`
	Rack      string    `json:"rack"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	inStock   float64   //实际库存量（即无需外购，最大值为Request）
}

func (bi BillItem) MarshalJSON() ([]byte, error) {
	type billitem BillItem
	return json.Marshal(struct {
		InStock float64 `json:"in_stock"`
		billitem
	}{bi.inStock, billitem(bi)})
}

func RemoveEmptyBills() {
	_, err := db.Exec(`DELETE FROM bom WHERE NOT id IN 
		(SELECT distinct bom_id FROM bom_item)`)
	assert(err)
}

func InventoryWIP() int {
	var ids []int
	assert(db.Select(&ids, `SELECT id FROM bom WHERE type=3 AND status=0`))
	if len(ids) == 0 {
		return 0
	}
	return ids[0]
}

func IsBillEmpty(bid int) bool {
	var cnt int
	assert(db.Get(&cnt, `SELECT COUNT(id) FROM bom_item where bom_id=?`, bid))
	return cnt == 0
}

func ListBillSummary(billType, uid int) []BillSummary {
	user := " "
	if uid > 0 {
		user = fmt.Sprintf(` AND user_id=%d`, uid)
	}
	qry := fmt.Sprintf(`SELECT COUNT(id) AS count,strftime('%%Y-%%m', created) AS
		month FROM bom WHERE type=%d%s GROUP BY month ORDER BY month DESC`, billType,
		user)
	var bs []BillSummary
	assert(db.Select(&bs, qry))
	return bs
}

func ListBills(billType, uid int, month string) (bills []Bill) {
	firstDay := month + "-01" //month格式为yyyy-mm
	start, _ := time.Parse("2006-01-02", firstDay)
	until := time.Date(start.Year(), start.Month()+1, start.Day(), 0, 0, 0, 0, time.Local)
	lastDay := until.Format("2006-01-02")
	user := ""
	if uid > 0 {
		user = fmt.Sprintf(` AND user_id=%d`, uid)
	}
	qry := fmt.Sprintf(`SELECT * FROM bom WHERE type=?%s AND created>=? 
		AND created<=?`, user)
	assert(db.Select(&bills, qry, billType, firstDay, lastDay))
	if len(bills) == 0 {
		return
	}
	bm := make(map[int]Bill)
	var ids []interface{}
	for _, b := range bills {
		bm[b.ID] = b
		ids = append(ids, b.ID)
	}
	var bis []BillItem
	assert(db.Select(&bis, `SELECT * FROM bom_item WHERE bom_id IN (?`+
		strings.Repeat(`,?`, len(ids)-1)+`)`, ids...))
	for _, bi := range bis {
		b := bm[bi.BomID]
		b.Count++
		if b.Type == 3 { //盘点单
			//TODO: 计算盘点单成本
		} else { //非盘点单
			if b.Status == 0 {
				b.Cost += math.Abs(bi.Cost * float64(bi.Request))
			} else {
				b.Cost += math.Abs(bi.Cost * float64(bi.Confirm))
			}
		}
		bm[bi.BomID] = b
	}
	bills = nil
	for _, b := range bm {
		bills = append(bills, b)
	}
	sort.Slice(bills, func(i, j int) (res bool) {
		bi := bills[i]
		bj := bills[j]
		diff := bi.Created.Unix() - bj.Created.Unix()
		if diff > 0 {
			return true
		}
		if diff < 0 {
			return false
		}
		return bi.ID > bj.ID
	})
	return
}

func GetBill(id int, itmOrd int) (bill Bill, items []BillItem) {
	assert(db.Get(&bill, `SELECT * FROM bom WHERE id=?`, id))
	qry := `SELECT bi.*,rack FROM bom_item bi,goods g WHERE gid=g.id AND
	    bom_id=? ORDER BY `
	switch itmOrd {
	case 0:
		qry += `bi.id DESC`
		assert(db.Select(&items, qry, id))
	case 1:
		qry += `g.rack`
		assert(db.Select(&items, qry, id))
	default: //除以上两种itmOrd外，不返回items
		return
	}
	var gs []Goods
	assert(db.Select(&gs, `SELECT id,stock FROM goods WHERE id IN (
		SELECT gid FROM bom_item WHERE bom_id=?)`, id))
	bill.Count = len(items)
	for i, it := range items {
		switch bill.Type {
		case 1: //入库单
			if bill.Status == 0 {
				bill.Cost += math.Abs(it.Cost * float64(it.Request))
			} else {
				bill.Cost += math.Abs(it.Cost * float64(it.Confirm))
			}
		case 2: //出库单
			it.Request = -it.Request
			it.Confirm = -it.Confirm
			it.inStock = func() float64 {
				if bill.Status > 0 { //已经出库，不再计算库存变化
					return it.Confirm * float64(bill.Sets)
				}
				var stock float64
				for _, g := range gs {
					if it.GoodsID == g.ID {
						stock = g.Stock
						break
					}
				}
				if stock > it.Request*float64(bill.Sets) {
					return it.Request * float64(bill.Sets)
				}
				return stock
			}()
			items[i] = it
			if it.Flag == 0 {
				if bill.Status == 0 {
					bill.Cost += math.Abs(it.Cost * float64(it.inStock) / float64(bill.Sets))
				} else {
					bill.Cost += math.Abs(it.Cost * float64(it.Confirm))
				}
			}
		case 3: //盘点单
			//TODO：计算盘点单的cost
			bill.Cost += math.Abs(it.Cost * float64(it.Request))
		}
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
	if b.Sets == 0 {
		b.Sets = 1
	}
	if b.ID == 0 {
		res := tx.MustExec(`INSERT INTO bom (type,user_id,markup,fee,memo,
			sets,paid) VALUES (?,?,?,?,?,?,?)`, b.Type, b.User, b.Markup,
			b.Fee, b.Memo, b.Sets, b.Paid)
		id, err := res.LastInsertId()
		assert(err)
		return int(id)
	}
	tx.MustExec(`UPDATE bom SET user_id=?,markup=?,fee=?,memo=?,sets=?,
		status=?,paid=?,courier=? WHERE ID=?`, b.User, b.Markup, b.Fee, b.Memo, b.Sets,
		b.Status, b.Paid, b.Courier, b.ID)
	return b.ID
}

func CloneBillItems(b Bill, ref int) {
	_, items := GetBill(ref, 0)
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	tx.MustExec(`DELETE FROM bom_item WHERE bom_id=?`, b.ID)
	for _, it := range items {
		tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,cost,request,confirm,memo) VALUES
	        (?,?,?,?,?,?,?)`, b.ID, it.GoodsID, it.GoodsName, it.Cost, -it.Request, 0, it.Memo)
	}
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
	tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,cost,request,confirm,flag,memo) VALUES 
		(?,?,?,?,?,?,?,?)`, bi.BomID, bi.GoodsID, bi.GoodsName, bi.Cost, bi.Request, bi.Confirm,
		bi.Flag, bi.Memo)
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

//请注意：不论是什么类型的Bill，只有在状态0转变为1的时候才会去修改库存！
//其它的状态修改可以调用SetBill，不能用这个函数。
func SetInventoryByBill(bid int) {
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
	if b.Status != 0 {
		panic(fmt.Errorf("bill#%d.status=%d, cannot set inventory", bid, b.Status))
	}
	tx.MustExec(`UPDATE bom SET status=? WHERE id=?`, 1, b.ID)
	switch b.Type {
	case 1:
		var bis []BillItem
		assert(tx.Select(&bis, `SELECT gid,confirm FROM bom_item WHERE bom_id=?`, b.ID))
		for _, bi := range bis {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, bi.Confirm, bi.GoodsID)
		}
	case 2:
		type billReq struct {
			ID        int
			Stock     float64
			Requested float64
			Flag      int
		}
		var br []billReq
		assert(tx.Select(&br, `SELECT g.id,g.stock,bi.flag,ABS(bi.request)*b.sets AS
			requested FROM goods g,bom_item bi,bom b WHERE b.id=bi.bom_id AND bi.gid=g.id
			AND	b.id=?`, bid))
		for _, r := range br {
			if r.Flag != 0 { //自备药材
				tx.MustExec(`UPDATE bom_item SET confirm=0 WHERE gid=? AND bom_id=?`, r.ID,
					bid)
			} else if r.Requested <= r.Stock {
				tx.MustExec(`UPDATE goods SET stock=stock-? WHERE id=?`, r.Requested, r.ID)
				tx.MustExec(`UPDATE bom_item SET confirm=request WHERE gid=? AND bom_id=?`,
					r.ID, bid)
			} else {
				confirm := -r.Stock / float64(b.Sets)
				tx.MustExec(`UPDATE goods SET stock=0 WHERE id=?`, r.ID)
				tx.MustExec(`UPDATE bom_item SET confirm=? WHERE gid=? AND bom_id=?`,
					confirm, r.ID, bid)
			}
		}
	case 3:
		tx.MustExec(`DELETE FROM bom_item WHERE flag=0 AND bom_id=?`, bid)
		var bis []BillItem
		assert(tx.Select(&bis, `SELECT gid,confirm FROM bom_item WHERE bom_id=?`, bid))
		for _, bi := range bis {
			tx.MustExec(`UPDATE goods SET stock=? WHERE id=?`, bi.Confirm, bi.GoodsID)
		}
	default:
		panic(fmt.Errorf("unsupported bill type %v", b.Type))
	}
}

func CreateInventory(uid int) int {
	id := SetBill(Bill{Type: 3, User: uid})
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
		tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,cost,request,confirm)
			VALUES (?,?,?,?,?,?)`, id, g.ID, g.Name, g.Cost, g.Stock, g.Stock)
	}
	return id
}
