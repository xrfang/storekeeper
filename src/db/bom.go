package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var pf [2]float64 //包装费用，全系统公用且运行时不可更改

func SetPackFee(large, small float64) {
	pf[0] = large
	pf[1] = small
}

/*
入库单状态：0=未完成；1=已锁单；2=已支付；3=已入库
出库单状态：0=未完成；1=已锁库；2=已出库；3=已收款
盘点单状态：0=进行中；1=已完成
总账单状态：0=进行中；1=已完成
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点；4=总帐
	User    int       `json:"user" db:"user_id"`
	Markup  float64   `json:"markup"`
	Fee     float64   `json:"fee"`
	Sets    int       `json:"sets"`
	Cost    float64   `json:"cost"`     //非数据库条目，实时计算，表示单剂药的成本
	Count   int       `json:"count"`    //非数据库条目，实时计算
	PackFee float64   `json:"pack_fee"` //非数据库条目，实时计算，包装袋总费用
	Memo    string    `json:"memo"`
	Status  int       `json:"status"`
	Paid    float64   `json:"paid"`
	Courier string    `json:"courier"`
	Ledger  int       `json:"ledger"`  //非总账单所属的总账单ID，若为0表示尚未计入总账单
	Changed int64     `json:"changed"` //status最后变化时间戳，注意：除status以外的属性变化不管！
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
	grp := `datetime(changed,'unixepoch','localtime')`
	if billType == 2 {
		grp = `created`
	}
	qry := fmt.Sprintf(`SELECT COUNT(id) AS count,strftime('%%Y-%%m', %s) AS
		month FROM bom WHERE type=%d%s GROUP BY month ORDER BY month DESC`,
		grp, billType, user)
	var bs []BillSummary
	assert(db.Select(&bs, qry))
	return bs
}

func ListBills(billType, uid int, month string) (bills []Bill) {
	firstDay := month + "-01" //month格式为yyyy-mm
	since, _ := time.Parse("2006-01-02", firstDay)
	until := time.Date(since.Year(), since.Month()+1, since.Day(), 0, 0, 0, 0, time.Local)
	user := ""
	if uid > 0 {
		user = fmt.Sprintf(` AND user_id=%d`, uid)
	}
	if billType == 2 {
		qry := fmt.Sprintf(`SELECT * FROM bom WHERE type=?%s AND created>=? 
		    AND created<=?`, user)
		assert(db.Select(&bills, qry, billType, since, until))
	} else {
		qry := fmt.Sprintf(`SELECT * FROM bom WHERE type=?%s AND changed>=? 
		    AND changed<=?`, user)
		assert(db.Select(&bills, qry, billType, since.Unix(), until.Unix()))
	}
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
	xp := make(map[int][2]float64) //先煎/后下药材的额外包装费用
	for _, bi := range bis {
		b := bm[bi.BomID]
		b.Count++
		x := xp[bi.BomID]
		switch b.Type {
		case 1: //入库单
			if b.Status == 3 { //已入库
				b.Cost += math.Abs(bi.Cost * float64(bi.Confirm))
			} else { //未入库
				b.Cost += math.Abs(bi.Cost * float64(bi.Request))
			}
		case 2: //出库单
			if strings.Contains(bi.Memo, "先煎") {
				x[0] = pf[1]
			}
			if strings.Contains(bi.Memo, "后下") {
				x[1] = pf[1]
			}
			if b.Status == 0 { //未锁库
				b.Cost += math.Abs(bi.Cost * float64(bi.Request))
			} else { //已锁库
				b.Cost += math.Abs(bi.Cost * float64(bi.Confirm))
			}
		case 3: //盘点单
			//TODO: 计算盘点单成本
		case 4: //总账单
		}
		bm[bi.BomID] = b
		xp[bi.BomID] = x
	}
	bills = nil
	for _, b := range bm {
		b.Cost = float64(int(b.Cost*10000)) / float64(10000)
		x := xp[b.ID]
		b.PackFee = float64(b.Sets) * (pf[0] + x[0] + x[1])
		bills = append(bills, b)
	}
	sort.Slice(bills, func(i, j int) (res bool) {
		bi := bills[i]
		bj := bills[j]
		diff := bi.Changed - bj.Changed
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
	case 2:
		qry += `g.pinyin`
		assert(db.Select(&items, qry, id))
	default: //除以上三种itmOrd外，不返回items
		return
	}
	var gs []Goods
	assert(db.Select(&gs, `SELECT id,stock FROM goods WHERE id IN (
		SELECT gid FROM bom_item WHERE bom_id=?)`, id))
	var xp [2]float64 //先煎/后下药材的额外包装费用，仅用于出库单
	bill.Count = len(items)
	for i, it := range items {
		switch bill.Type {
		case 1: //入库单
			if bill.Status == 3 { //已入库
				bill.Cost += math.Abs(it.Cost * float64(it.Confirm))
			} else { //未入库
				bill.Cost += math.Abs(it.Cost * float64(it.Request))
			}
		case 2: //出库单
			if strings.Contains(it.Memo, "先煎") {
				xp[0] = pf[1]
			}
			if strings.Contains(it.Memo, "后下") {
				xp[1] = pf[1]
			}
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
	bill.PackFee = float64(bill.Sets) * (pf[0] + xp[0] + xp[1]) //仅用于出库单
	bill.Cost = float64(int(bill.Cost*10000)) / float64(10000)
	return
}

func FindBillItem(bid int, item string) (items []BillItem) {
	item = strings.ToUpper(item)
	assert(db.Select(&items, `SELECT * FROM bom_item WHERE bom_id=?
		AND gid IN (SELECT id FROM goods WHERE name=? OR pinyin=?)`,
		bid, item, item))
	return
}

func GetBillItems(bid int, gid ...interface{}) (items []BillItem) {
	ids := []interface{}{bid}
	qry := `SELECT * FROM bom_item WHERE bom_id=?`
	if len(gid) > 0 {
		ids = append(ids, gid...)
		qry += ` AND gid IN (?` + strings.Repeat(`,?`, len(gid)-1) + `)`
	}
	assert(db.Select(&items, qry, ids...))
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
	currentStatus := 0
	if b.ID != 0 {
		err := tx.Get(&currentStatus, `SELECT status FROM bom WHERE id=?`, b.ID)
		if err != nil {
			panic(fmt.Errorf("status(#%v): %v", b.ID, err))
		}
	}
	now := time.Now().Unix()
	props := make(map[string]interface{})
	switch currentStatus {
	case 0:
		if b.ID == 0 {
			props["type"] = b.Type
			props["changed"] = now
		}
		props["user_id"] = b.User
		props["markup"] = b.Markup
		props["fee"] = b.Fee
		if b.Sets <= 0 {
			props["sets"] = 1
		} else {
			props["sets"] = b.Sets
		}
		props["memo"] = b.Memo
	case 1:
		props["user_id"] = b.User
		props["markup"] = b.Markup
		props["fee"] = b.Fee
		props["memo"] = b.Memo
		props["courier"] = b.Courier
		if b.Type == 1 {
			props["paid"] = b.Paid
		}
	case 2:
		props["user_id"] = b.User
		props["markup"] = b.Markup
		props["fee"] = b.Fee
		props["memo"] = b.Memo
		props["courier"] = b.Courier
		props["paid"] = b.Paid
	case 3: //理论上说，处于此状态的订单可以修改ledger属性（即所属总账单ID）
		//但是，因为有专门的流程做总账单生成，因此不允许在此设置
	default:
		panic(fmt.Errorf("not changeable: status(#%v)=%v", b.ID, currentStatus))
	}
	if currentStatus != b.Status {
		props["status"] = b.Status
		props["changed"] = now
	}
	if b.ID == 0 {
		t, ok := props["type"].(byte)
		if !ok || t < 1 || t > 3 {
			panic(errors.New("missing or invalid bill type"))
		}
		uid, _ := props["user_id"].(int)
		var c int
		tx.Get(&c, `SELECT COUNT(id) FROM user WHERE id=?`, uid)
		if c != 1 {
			panic(errors.New("missing or invalid user_id"))
		}
		var ks []string
		var vs []interface{}
		for k, v := range props {
			ks = append(ks, k)
			vs = append(vs, v)
		}
		cmd := fmt.Sprintf(`INSERT INTO bom (%s) VALUES (%s)`, strings.Join(ks, ","),
			strings.Repeat("?,", len(vs)-1)+`?`)
		res := tx.MustExec(cmd, vs...)
		id, err := res.LastInsertId()
		assert(err)
		return int(id)
	}
	var ks []string
	var vs []interface{}
	for k, v := range props {
		ks = append(ks, fmt.Sprintf(`%s=?`, k))
		vs = append(vs, v)
	}
	vs = append(vs, b.ID)
	cmd := fmt.Sprintf(`UPDATE bom SET %s WHERE id=?`, strings.Join(ks, ","))
	tx.MustExec(cmd, vs...)
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

/*
针对不同类型单据，本函数适用性如下：
- 进货单在状态2=>3的时候变更库存，因此，只能设置状态为2的进货单
- 出货单再状态0=>1的时候变更库存，因此，只能设置状态为0的出货单
- 盘点单与出货单相同状态条件
- 总账单不适用本函数
【请注意】其它的状态修改可以调用SetBill，不能用这个函数。
*/
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
	now := time.Now().Unix()
	switch b.Type {
	case 1:
		if b.Status != 2 {
			panic(fmt.Errorf("bill#%d.status=%d, cannot set inventory", bid, b.Status))
		}
		tx.MustExec(`UPDATE bom SET status=?,changed=? WHERE id=?`, 3, now, b.ID)
		var bis []BillItem
		assert(tx.Select(&bis, `SELECT gid,confirm FROM bom_item WHERE bom_id=?`, b.ID))
		for _, bi := range bis {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, bi.Confirm, bi.GoodsID)
		}
	case 2:
		if b.Status != 0 {
			panic(fmt.Errorf("bill#%d.status=%d, cannot set inventory", bid, b.Status))
		}
		tx.MustExec(`UPDATE bom SET status=?,changed=? WHERE id=?`, 1, now, b.ID)
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
		if b.Status != 0 {
			panic(fmt.Errorf("bill#%d.status=%d, cannot set inventory", bid, b.Status))
		}
		tx.MustExec(`UPDATE bom SET status=?,changed=? WHERE id=?`, 1, now, b.ID)
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
