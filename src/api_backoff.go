package main

import (
	"fmt"
	"net/http"
	"regexp"
	"storekeeper/db"
	"strconv"
	"strings"
)

func jr(w http.ResponseWriter, stat bool, arg interface{}) {
	rep := map[string]interface{}{"stat": stat}
	if arg != nil {
		if stat {
			rep["data"] = arg
		} else {
			rep["mesg"] = fmt.Sprintf("%v", arg)
		}
	}
	jsonReply(w, rep)
}

func apiBackOffice(cf Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ps := strings.Split(r.URL.Path[len(cf.BackOff)+2:], "/")
		if len(ps) < 1 {
			mesg := fmt.Sprintf("USAGE: /%s/<target>/...", cf.BackOff)
			mesg += "\n\nvalid targets are: db,user,goods,bom,bom_item"
			jr(w, false, mesg)
			return
		}
		target := ps[0]
		action := ""
		if len(ps) > 1 {
			action = ps[1]
		}
		params := []string{}
		if len(ps) > 2 {
			params = ps[2:]
		}
		args := r.URL.Query()
		switch target {
		case "db": //数据库查询（不支持曾删改）
			if action == "" {
				jr(w, false, "missing table name")
				return
			}
			if action != "user" && action != "goods" && action != "bom" &&
				action != "bom_item" {
				jr(w, false, "table name must be: user, goods, bom or bom_item")
				return
			}
			r := regexp.MustCompile(`^\w+$`)
			var (
				ck     []string
				op     []string
				cv     []interface{}
				ok, ov string
				lim    int
			)
			for k := range args {
				if k == "@" {
					vs := strings.SplitN(args.Get("@"), ":", 2)
					if len(vs[0]) > 0 {
						if vs[0][0] == '!' {
							ok = vs[0][1:]
							ov = "DESC"
						} else {
							ok = vs[0]
							ov = "ASC"
						}
					}
					if len(vs) == 2 {
						lim, _ = strconv.Atoi(vs[1])
					}
					continue
				}
				if !r.MatchString(k) {
					jr(w, false, "invalid argument: "+k)
					return
				}
				ck = append(ck, k)
				v := strings.SplitN(strings.ToUpper(args.Get(k)), ":", 2)
				if len(v) == 1 {
					op = append(op, "=")
					cv = append(cv, v[0])
				} else {
					switch v[0] {
					case "=", "<", ">", "<=", ">=", "<>":
						op = append(op, v[0])
						cv = append(cv, v[1])
					case "~":
						op = append(op, "LIKE")
						s := strings.ReplaceAll(v[1], "*", "%")
						if s == v[1] { //没有通配符
							s = "%" + s + "%" //则认为两边通配
						}
						cv = append(cv, s)
					case "(":
						vs := strings.Split(v[1], ",")
						op = append(op, fmt.Sprintf("IN%d", len(vs)))
						for _, v := range vs {
							cv = append(cv, v)
						}
					default:
						jr(w, false, "invalid operator: "+v[0])
						return
					}
				}
			}
			qry := fmt.Sprintf(`SELECT * FROM %s`, action)
			if len(ck) > 0 {
				var where []string
				for i, k := range ck {
					var cond string
					if strings.HasPrefix(op[i], "IN") {
						c, _ := strconv.Atoi(op[i][2:])
						cond = fmt.Sprintf(`"%s" IN (?%s)`, k,
							strings.Repeat(",?", c-1))
					} else {
						cond = fmt.Sprintf(`"%s" %s ?`, k, op[i])
					}
					where = append(where, cond)
				}
				qry += ` WHERE ` + strings.Join(where, " AND ")
			}
			if ok != "" {
				qry += fmt.Sprintf(" ORDER BY %s %s", ok, ov)
			}
			if lim > 0 {
				qry += fmt.Sprintf(" LIMIT %d", lim)
			}
			res, err := db.RawSelect(qry, cv)
			if err == nil {
				jr(w, true, res)
			} else {
				jr(w, false, err)
			}
		case "bom":
			switch action {
			case "user": //修改入库单用户
				val, err := db.BomSetUser(params)
				if err == nil {
					jr(w, true, val)
				} else {
					jr(w, false, err)
				}
			case "paid": //修改入库单实付金额
				val, err := db.BomSetPaid(params)
				if err == nil {
					jr(w, true, val)
				} else {
					jr(w, false, err)
				}
			case "set": //已经锁库以后修改出库单抓药剂数
				val, err := db.BomSetAmount(params)
				if err == nil {
					jr(w, true, val)
				} else {
					jr(w, false, err)
				}
			case "del": //删除出库单
				err := db.BomDelete(params)
				if err == nil {
					jr(w, true, nil)
				} else {
					jr(w, false, err)
				}
			case "item":
				val, err := db.BomSetItem(params, args)
				if err == nil {
					jr(w, true, val)
				} else {
					jr(w, false, err)
				}
			case "memo":
				err := db.BomSetMemo(params)
				if err == nil {
					jr(w, true, nil)
				} else {
					jr(w, false, err)
				}
			case "help":
				jr(w, true, map[string]interface{}{
					"user": "修改入库单用户",
					"paid": "修改入库单实付金额",
					"set":  "修改出库单抓药剂数",
					"item": "增删改出库单条目",
					"memo": "修改已关订单备注",
					"del":  "删除出库单",
				})
			default:
				mesg := fmt.Sprintf("invalid action, try `/%s/bom/help`", cf.BackOff)
				jr(w, false, mesg)
			}
		default:
			jr(w, false, "Invalid target, expect: db,bom")
		}
	}
}
