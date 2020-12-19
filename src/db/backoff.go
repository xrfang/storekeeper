package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
			//err = trace("%v", e)
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

func BomSetAmount(params []string) (interface{}, error) {
	/*
	   已经锁库以后修改抓药剂数
	   - 仅针对出库单
	   - 状态为锁库、出货（不能是未锁库或已关闭）
	*/
	return nil, nil
}

func BomDelete(params []string) error {
	/*
	  - 仅针对出库单
	  - 状态为锁库
	*/
	return nil
}

func BomItemAdd(parms []string) (interface{}, error) { //增加一味药材
	return nil, nil
}

func BomItemDel(parms []string) (interface{}, error) { //删除一味药材
	return nil, nil
}

func BomItemAlt(parms []string) (interface{}, error) { //调整某药材用量
	return nil, nil
}

func BomItemGet(parms []string) (interface{}, error) { //（有新入库后）继续抓药
	return nil, nil
}
