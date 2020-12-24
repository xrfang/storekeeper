package db

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func LedgerList(params []string) (ret interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	type ledgerInfo struct {
		ID      string `json:"id"`
		Status  int    `json:"status"`
		Created int64  `json:"created"`
		Changed int64  `json:"changed"`
	}
	var lis []ledgerInfo
	assert(db.Select(&lis, `SELECT id,status,strftime("%s",created) 
		AS created,changed FROM bom WHERE type=4`))
	var res []map[string]interface{}
	for _, li := range lis {
		res = append(res, map[string]interface{}{
			"id":      li.ID,
			"status":  li.Status,
			"created": time.Unix(li.Created, 0),
			"changed": time.Unix(li.Changed, 0),
		})
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
		tx.MustExec(`DELETE FROM bom WHERE id=?`, id)
		err = errors.New("no order eligible for new ledger")
	}
	return
}

func LedgerGet(params []string) (ret interface{}, err error) {
	return
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
	res := db.MustExec(`UPDATE bom SET status=1,changed=? WHERE type=4 
	    AND status=0 AND id=?`, time.Now().Unix(), lid)
	ra, err := res.RowsAffected()
	assert(err)
	if ra == 0 {
		panic(fmt.Errorf("ledger#%d: not found or already closed", lid))
	}
	return
}
