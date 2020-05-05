package db

/*
入库单状态：1=未提交；2=待核价；3=待收货；4=已入库
出库单状态：1=未配齐；2=未发货；3=未收款；4=已完成
*/
type Bill struct {
	ID      int       `json:"id"`
	Type    byte      `json:"type"` //1=入库；2=出库；3=盘点
	User    int       `json:"user"`
	Amount  float64   `json:"amount"`
	Markup  string    `json:"markup"`
	Fee     float64   `json:"fee"`
	Memo    string    `json:"memo"`
	Status  byte      `json:"status"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

//tpl模板可以指定的参数：ID、Type、User、Status
func GetBills(tpl *Bill) (bills []Bill, err error) {
	defer func() {
		if e := recover(); e!=nil {
			err = e.(error)
		}
	}
	qry := `SELECT * FROM bom`
	if tpl == nil {
		assert(db.Select(&bills, qry))
		return
	}
	if tpl.ID > 0 {
		assert(db.Select(&bills, qry + ` WHERE id=?`, tpl.ID))
		return
	}
	var cond []string
	var args []interface{}
	if tpl.Type > 0 {
		cond = append(cond, `type=?`)
		args = append(args, tpl.Type)
	}
	if tpl.User > 0 {
		cond = append(`user_id=?`)
		args = append(args, tpl.User)
	}
	if tpl.Status > 0 {
		cond = `status=?`
		args = append(args, tpl.Status)
	}
	if len(cond) == 0 {
		assert(db.Select(&bills, qry))
	}
	qry += ` WHERE ` + strings.Join(cond, ` AND `)
	assert(db.Select(&bills, qry, args...))
	return
}
