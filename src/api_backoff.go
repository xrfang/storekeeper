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
	if stat {
		jsonReply(w, map[string]interface{}{"stat": true, "data": arg})
	} else {
		jsonReply(w, map[string]interface{}{"stat": false, "mesg": arg})
	}
}

func apiBackOffice(cf Configuration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ps := strings.Split(r.URL.Path[len(cf.BackOff)+2:], "/")
		if len(ps) < 1 {
			mesg := fmt.Sprintf("USAGE: /%s/<target>/<action>/<params>?<args>", cf.BackOff)
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
		case "db":
			switch len(params) {
			case 0:
				jr(w, false, "missing table name")
				return
			case 1:
				t := strings.TrimSpace(params[0])
				if t != "user" && t != "goods" && t != "bom" && t != "bom_item" {
					jr(w, false, "table name must be: user, goods, bom or bom_item")
					return
				}
				switch action {
				case "select":
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
						v := strings.SplitN(args.Get(k), ":", 2)
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
								cv = append(cv, strings.ReplaceAll(v[1], "*", "%"))
							default:
								jr(w, false, "invalid operator: "+v[0])
								return
							}
						}
					}
					qry := fmt.Sprintf(`SELECT * FROM %s`, t)
					if len(ck) > 0 {
						var where []string
						for i, k := range ck {
							where = append(where, fmt.Sprintf(`'%s' %s ?`, k, op[i]))
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
				case "insert":
					jr(w, false, "INSERT not implemented yet")
				case "update":
				default:
					jr(w, false, "db action must be `select`,`insert` or `update`")
				}
			default:
				jr(w, false, "excessive params")
				return
			}
		case "user":
		case "goods":
		case "bom":
		case "bom_item":
		default:
			jr(w, false, "Invalid target, expect: db,user,goods,bom,bom_item")
		}
	}
}
