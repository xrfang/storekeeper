package db

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type (
	ledgerInfo struct {
		ID      int   `json:"id"`
		Status  int   `json:"status"`
		Created int64 `json:"created"`
		Changed int64 `json:"changed"`
	}
	bInfo struct {
		BID    int
		UID    int
		Client int
		Markup float64
		Extra  float64
		Cost   float64
		Pack   float64
		Paid   float64
		Status int
	}
	bSummary struct {
		Goods float64 `json:"goods"`
		Extra float64 `json:"fees"`
		Pack  float64 `json:"package"`
		Items []int   `json:"items"`
	}
)

func (s bSummary) Round() bSummary {
	s.Extra = math.Round(s.Extra*100) / 100
	s.Goods = math.Round(s.Goods*100) / 100
	s.Pack = math.Round(s.Pack*100) / 100
	return s
}

func (li ledgerInfo) Export() map[string]interface{} {
	return map[string]interface{}{
		"id":      li.ID,
		"status":  li.Status,
		"created": time.Unix(li.Created, 0),
		"changed": time.Unix(li.Changed, 0),
	}
}

func LedgerList(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	var lis []ledgerInfo
	assert(db.Select(&lis, `SELECT id,status,strftime("%s",created) 
		AS created,changed FROM bom WHERE type=4`))
	var res []map[string]interface{}
	for _, li := range lis {
		res = append(res, li.Export())
	}
	return res, nil
}

func LedgerNew() (id int64, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	tx := db.MustBegin()
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	res := tx.MustExec(`INSERT INTO bom (type,user_id,changed) 
	    VALUES (4,0,?)`, time.Now().Unix()) //总账单的user_id一律设为0
	id, err = res.LastInsertId()
	assert(err)
	res = tx.MustExec(`UPDATE bom SET ledger=? WHERE type IN (1,2) AND 
		status>=2 AND ledger=0`, id)
	if ra, _ := res.RowsAffected(); ra == 0 {
		panic(errors.New("no order eligible for new ledger"))
	}
	type InValue struct {
		Count int
		Value float64
	}
	var iv InValue
	assert(tx.Get(&iv, `SELECT COUNT(id) AS count,SUM(cost*stock) AS
		value FROM goods WHERE stock>0`))
	tx.MustExec(`INSERT INTO bom_item (bom_id,gid,gname,cost,request,confirm)
	    VALUES (?,0,'药材统计',?,?,?)`, id, int(iv.Value), iv.Count, iv.Count)
	return
}

func ledgerGetCheckin(id int) (ret interface{}) {
	rows, err := db.Queryx(`SELECT b.id,u.name,b.paid*1.0 AS paid,CAST(strftime(
		"%s",b.created) AS int) AS created,changed,b.memo FROM bom b, user u WHERE 
		b.user_id=u.id AND type=1 AND ledger=?`, id)
	assert(err)
	total := make(map[string]float64)
	var list []map[string]interface{}
	for rows.Next() {
		l := make(map[string]interface{})
		assert(rows.MapScan(l))
		l["created"] = time.Unix(l["created"].(int64), 0)
		l["changed"] = time.Unix(l["changed"].(int64), 0)
		list = append(list, l)
		name := l["name"].(string)
		total[name] += l["paid"].(float64)
	}
	assert(rows.Err())
	ret = map[string]interface{}{"total": total, "bills": list}
	return
}

func summarize(bis []bInfo) bSummary {
	var bs bSummary
	for _, bi := range bis {
		bs.Items = append(bs.Items, bi.BID)
		bs.Extra += bi.Extra
		bi.Paid -= bi.Extra
		cost := bi.Cost
		if bi.Markup < 0 {
			cost *= sm
		} else {
			cost *= 1 + bi.Markup/100
		}
		bs.Goods += cost
		bi.Paid -= cost
		if bi.Paid > 0 {
			bs.Pack += bi.Paid
		} else {
			bs.Goods -= bi.Paid
		}
	}
	return bs.Round()
}

func ledgerGetCheckout(id int) (ret interface{}) {
	var us []User
	assert(db.Select(&us, `SELECT * FROM user`))
	um := make(map[int]string)
	for _, u := range us {
		um[u.ID] = u.Name
	}
	var bis []bInfo
	assert(db.Select(&bis, `SELECT b.id AS bid,u.id AS uid,u.client,u.markup,
		b.paid,b.status,b.fee AS extra FROM bom b, user u WHERE u.id=b.user_id
		AND type=2 AND status IN (2,3) AND ledger=?`, id))
	var done []bInfo
	todo := make(map[string][]bInfo)
	for _, bi := range bis {
		b, _ := GetBill(bi.BID, 0)
		cost := b.Cost * float64(b.Sets)
		bi.Cost = cost
		bi.Pack = b.PackFee
		if bi.Status == 2 {
			uid := bi.UID
			if bi.Client != 0 {
				if bi.Markup == 0 {
					uid = bi.Client //主账号
				} else {
					uid = 0 //其他人
				}
			}
			if uid == 0 {
				todo["其他人"] = append(todo["其他人"], bi)
			} else {
				todo[um[uid]] = append(todo[um[uid]], bi)
			}
		} else {
			done = append(done, bi)
		}
	}
	pending := make(map[string]bSummary)
	for name, bis := range todo {
		pending[name] = summarize(bis)
	}
	return map[string]interface{}{
		"received": summarize(done),
		"pending":  pending,
	}
}

func ledgerGetInventory(id int) (ret interface{}) {
	it := GetBillItems(id)[0]
	return map[string]interface{}{
		"cost":  it.Cost,
		"count": it.Request,
		"time":  it.Created,
	}
}

func LedgerGet(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 1 {
		panic(errors.New("bad format, use /ledger/get/<ledger_id>"))
	}
	var li ledgerInfo
	assert(db.Get(&li, `SELECT id,status,strftime("%s",created) 
		AS created,changed FROM bom WHERE type=4 AND id=?`, params[0]))
	return map[string]interface{}{
		"ledger":    li.Export(),
		"inventory": ledgerGetInventory(li.ID),
		"checkin":   ledgerGetCheckin(li.ID),
		"checkout":  ledgerGetCheckout(li.ID),
	}, nil
}

func LedgerDel(params []string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 1 {
		panic(errors.New("bad format, use /ledger/del/<ledger_id>"))
	}
	lid, _ := strconv.Atoi(params[0])
	tx := db.MustBegin()
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	var status int
	assert(tx.Get(&status, `SELECT status FROM bom WHERE type=4 AND id=?`, lid))
	if status > 0 {
		panic(errors.New("cannot delete closed ledger"))
	}
	tx.MustExec(`DELETE FROM bom_item WHERE bom_id=?`, lid)
	tx.MustExec(`UPDATE bom SET ledger=0 WHERE ledger=?`, lid)
	tx.MustExec(`DELETE FROM bom WHERE id=?`, lid)
	return
}

func LedgerCls(params []string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 1 {
		panic(errors.New("bad format, use /ledger/cls/<ledger_id>"))
	}
	lid, _ := strconv.Atoi(params[0])
	panic(errors.New("TODO: 设置paid属性")) //仅对内部订单，检查不允许有未关闭的外部订单
	res := db.MustExec(`UPDATE bom SET status=1,changed=? WHERE type=4 
	    AND status=0 AND id=?`, time.Now().Unix(), lid)
	ra, err := res.RowsAffected()
	assert(err)
	if ra == 0 {
		panic(fmt.Errorf("ledger#%d: not found or already closed", lid))
	}
	return
}
