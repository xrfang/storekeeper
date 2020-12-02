package db

import (
	"errors"
	"time"
)

type Ledger struct {
}

func GetLedger(lid int) *Ledger {
	return nil //TODO
}

//创建一个新的总账单，将当前完成而又没有加入总账单的订单加入该账单
func LedgerBills() int64 {
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	res := tx.MustExec(`INSERT INTO bom (type,user_id,changed) 
	    VALUES (4,0,?)`, time.Now().Unix()) //总账单的user_id一律设为0
	lid, err := res.LastInsertId()
	assert(err)
	tx.MustExec(`UPDATE bom SET ledger=? WHERE type IN (1,2) AND 
	    status>=2 AND ledger=0`, lid)
	return lid
}

//删除一个总账单（只有未关闭的总账单，即status=0时才可以删除）
func UnledgerBills(lid int) {
	tx, err := db.Beginx()
	assert(err)
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback()
			panic(e)
		}
		assert(tx.Commit())
	}()
	var status int
	assert(tx.Get(&status, `SELECT status FROM bom WHERE id=?`, lid))
	if status > 0 {
		panic(errors.New("cannot delete closed ledger"))
	}
	tx.MustExec(`UPDATE bom SET ledger=0 WHERE ledger=?`, lid)
}
