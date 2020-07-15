package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("USAGE: %s <db-file>\n", filepath.Base(os.Args[0]))
		return
	}
	var err error
	db, err = sqlx.Connect("sqlite3", "file:"+os.Args[1]+"?cache=shared")
	assert(err)
	je := json.NewEncoder(os.Stdout)
	je.SetIndent("", "    ")
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		text, _ := r.ReadString('\n')
		ps := GetPSItems(text)
		fmt.Printf("got %d items\n", len(ps))
		je.Encode(ps)
		rx := GetPrevRx(ps)
		fmt.Printf("got %d prescriptions\n", len(rx))
		je.Encode(rx)
	}
}
