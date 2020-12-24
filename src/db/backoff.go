package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
)

func RawSelect(qry string, args []interface{}) (res []map[string]interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = trace("%v", e)
		}
	}()
	rows, err := db.Queryx(qry, args...)
	assert(err)
	for rows.Next() {
		r := make(map[string]interface{})
		assert(rows.MapScan(r))
		res = append(res, r)
	}
	return
}

func checkBOM(id string, btypes []int, stats []int) Bill {
	type tss struct {
		Type   int
		Status int
		Ledger int
	}
	var b Bill
	err := db.Get(&b, `SELECT * FROM bom WHERE id=?`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			panic(fmt.Errorf("unknown bom-id: %s", id))
		} else {
			panic(fmt.Errorf("checkBOM: %v", err))
		}
	}
	tm := false
	for _, t := range btypes {
		if int(b.Type) == t {
			tm = true
			break
		}
	}
	if !tm {
		panic(fmt.Errorf("bom#%s: invalid type %v", id, b.Type))
	}
	sm := false
	for _, s := range stats {
		if b.Status == s {
			sm = true
			break
		}
	}
	if !sm {
		panic(fmt.Errorf("bom#%s: invalid status %v", id, b.Status))
	}
	if b.Ledger != 0 {
		panic(fmt.Errorf("bom#%s: added to ledger#%d", id, b.Ledger))
	}
	return b
}

func BomSetUser(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 2 {
		panic(errors.New("bad format, use /bom/user/<bom_id>/<uid>"))
	}
	b := checkBOM(params[0], []int{1}, []int{1, 2, 3})
	user, err := strconv.Atoi(params[1])
	if err != nil || user <= 0 {
		panic(fmt.Errorf("invalid user-id: '%s'", params[1]))
	}
	var client int
	err = db.Get(&client, "SELECT client FROM user WHERE id=?", user)
	if err != nil {
		if err == sql.ErrNoRows {
			panic(fmt.Errorf("user#%d: not found", user))
		} else {
			panic(fmt.Errorf("checkUser: %v", err))
		}
	}
	if client != 0 {
		panic(fmt.Errorf("user#%d: not primary user", user))
	}
	if b.User == user {
		return nil, errors.New("user of bill not changed")
	}
	db.MustExec("UPDATE bom SET user_id=? WHERE id=?", user, b.ID)
	ret = map[string]interface{}{"old": b.User, "new": user}
	return
}

//修改实付金额（临时使用），仅针对入库单且为关单状态
func BomSetPaid(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 2 {
		panic(errors.New("bad format, use /bom/paid/<bom_id>/<amount>"))
	}
	b := checkBOM(params[0], []int{1}, []int{3})
	paid, err := strconv.ParseFloat(params[1], 64)
	if err != nil || paid < 0 {
		panic(fmt.Errorf("invalid paid amount: '%s'", params[1]))
	}
	if b.Paid == paid {
		return nil, errors.New("paid of bill not changed")
	}
	db.MustExec("UPDATE bom SET paid=? WHERE id=?", paid, b.ID)
	ret = map[string]interface{}{"old": b.Paid, "new": paid}
	return
}

func BomSetAmount(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 2 {
		panic(errors.New("bad format, use /bom/set/<bom_id>/<sets>"))
	}
	b := checkBOM(params[0], []int{2}, []int{1, 2})
	sets, err := strconv.Atoi(params[1])
	if err != nil || sets <= 0 || sets > 20 {
		panic(fmt.Errorf("invalid sets '%s' (must between 1~20)", params[1]))
	}
	if sets == b.Sets {
		return nil, errors.New("sets of bill not changed")
	}
	items := GetBillItems(b.ID)
	tx := db.MustBegin()
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
		SetInventoryByBill(b.ID)
		if b.Status == 2 {
			db.MustExec(`UPDATE bom SET status=2 WHERE id=?`, b.ID)
		}
	}()
	for _, it := range items {
		c := math.Abs(it.Confirm) * float64(b.Sets)
		if c > 0 {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, c, it.GoodsID)
		}
		tx.MustExec(`UPDATE bom_item SET confirm=0 WHERE id=?`, it.ID)
	}
	tx.MustExec(`UPDATE bom SET status=0,sets=? WHERE id=?`, sets, b.ID)
	return
}

func BomSetItem(params []string, args url.Values) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) < 2 || len(params) > 4 {
		panic(errors.New("bad format, use /bom/item/<bom_id>/<gname or pinyin>[/request][?flag,memo=...]"))
	}
	var (
		act   bool //是否需要执行动作
		bi    *BillItem
		g     *Goods
		cost  float64
		gid   int
		gname string
		flag  *int
		memo  *string
		req   *float64
	)
	//检查BOM
	b := checkBOM(params[0], []int{2}, []int{1, 2, 3})
	its := FindBillItem(b.ID, params[1])
	switch len(its) {
	case 0: //输入的药材是本方中没有的：需要增加该药材
		var gs []*Goods
		term := strings.ToUpper(params[1])
		assert(db.Select(&gs, `SELECT * FROM goods WHERE name=? OR pinyin=?`, term, term))
		switch len(gs) {
		case 0:
			panic(fmt.Errorf("no goods named '%s'", params[1]))
		case 1:
			g = gs[0]
		default:
			var ns []string
			for _, g := range gs {
				ns = append(ns, g.Name)
			}
			panic(fmt.Errorf("'%s' ambiguous: %s", params[1], strings.Join(ns, ", ")))
		}
	case 1: //输入的药材在本方中找到了：需要修改或删除该药材
		bi = &its[0]
		gid = bi.GoodsID
		gname = bi.GoodsName
	default:
		panic(fmt.Errorf("bom#%d: '%s' ambiguous", b.ID, params[1]))
	}
	//收集需要设置的参数
	if len(params) > 2 {
		req = new(float64)
		*req, err = strconv.ParseFloat(params[2], 64)
		if err != nil || *req < 0 {
			panic(fmt.Errorf("invalid request amount '%s'", params[2]))
		}
		*req = -math.Abs(*req)
		act = true
	}
	v, ok := args["flag"]
	if ok && len(v) > 0 {
		flag = new(int)
		*flag, err = strconv.Atoi(v[0])
		if err != nil || *flag < 0 || *flag > 1 {
			panic(fmt.Errorf("invalid flag '%s'", v[0]))
		}
		act = true
	}
	v, ok = args["memo"]
	if ok {
		memo = new(string)
		if len(v) > 0 {
			*memo = v[0]
		}
		act = true
	}
	if !act {
		panic(errors.New("nothing to do"))
	}
	//按照BOM状态处理修改工作
	if b.Status == 3 { //在状态3的时候只能修改memo
		if memo == nil {
			panic(fmt.Errorf("bom#%d: status=3, memo not provided", b.ID))
		}
		db.MustExec(`UPDATE bom_item SET memo=? WHERE id=?`, *memo, bi.ID)
		ret = map[string]interface{}{"old": bi.Memo, "new": *memo}
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
	var keys []string
	var vals []interface{}
	if bi == nil { //增加新药材
		if req == nil {
			panic(fmt.Errorf("bom#%d: invalid request amount for '%s'", b.ID, g.Name))
		}
		if flag == nil {
			flag = new(int)
			*flag = 0
		}
		gid = g.ID
		gname = g.Name
		cost = g.Cost
	} else { //修改或删除药材
		if req == nil { //没有提供request，使用原来的值
			req = new(float64)
			*req = bi.Request
		}
		if flag == nil {
			flag = new(int)
			*flag = bi.Flag
		}
		gid = bi.GoodsID
		gname = bi.GoodsName
		cost = bi.Cost
		c := math.Abs(bi.Confirm) * float64(b.Sets)
		if c > 0 {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, c, bi.GoodsID)
		}
		tx.MustExec(`DELETE FROM bom_item WHERE id=?`, bi.ID)
	}
	keys = []string{"bom_id", "gid", "gname", "cost", "request", "confirm", "flag"}
	vals = []interface{}{b.ID, gid, gname, cost, *req, 0, *flag}
	if memo != nil {
		keys = append(keys, "memo")
		vals = append(vals, *memo)
	}
	ret = map[string]interface{}{"old": bi, "new": map[string]interface{}{
		"keys": keys, "vals": vals}}
	cmd := fmt.Sprintf(`INSERT INTO bom_item (%s) VALUES (%s)`,
		strings.Join(keys, ","), `?`+strings.Repeat(`,?`, len(keys)-1))
	res := tx.MustExec(cmd, vals...)
	lid, err := res.LastInsertId()
	assert(err)
	if *flag == 0 { //非自备，需要扣减库存
		want := math.Abs(*req * float64(b.Sets))
		var have float64
		assert(tx.Get(&have, `SELECT stock FROM goods WHERE id=?`, gid))
		if have < want {
			want = have
			*req = -have / float64(b.Sets)
		}
		tx.MustExec(`UPDATE bom_item SET confirm=? WHERE id=?`, *req, lid)
		tx.MustExec(`UPDATE goods SET stock=stock-? WHERE id=?`, want, gid)
		assert(tx.Get(&have, `SELECT stock FROM goods WHERE id=?`, gid))
		if have < 0 {
			panic(errors.New("concurrent stock modification, try again"))
		}
		vals[5] = *req //设置一下confirm，仅供返回数据使用
	}
	return
}

func BomDelete(params []string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if len(params) != 1 {
		panic(errors.New("bad format, use /bom/del/<bom_id>"))
	}
	b := checkBOM(params[0], []int{2}, []int{1})
	items := GetBillItems(b.ID)
	tx := db.MustBegin()
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	for _, it := range items {
		c := math.Abs(it.Confirm) * float64(b.Sets)
		if c > 0 {
			tx.MustExec(`UPDATE goods SET stock=stock+? WHERE id=?`, c, it.GoodsID)
		}
		tx.MustExec(`DELETE FROM bom_item WHERE id=?`, it.ID)
	}
	tx.MustExec("DELETE FROM bom WHERE id=?", b.ID)
	return
}
