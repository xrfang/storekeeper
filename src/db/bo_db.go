package db

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
