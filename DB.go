package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

func InitDB() {
	db, err := sql.Open("sqlite3", "./SMS.db")
	if err != nil {
		log.Fatal(err)
	}
	SQLQuery := `CREATE TABLE IF NOT EXISTS "SMS" ("Device" TEXT,"Tittle" TEXT, "Data" TEXT);`
	db.Exec(SQLQuery)
	db.Close()
}

func AddSMSToDB(DeviceName string, Tittle string, Data string) {
	db, err := sql.Open("sqlite3", "./SMS.db")
	if err != nil {
		log.Fatal(err)
	}
	SQLQuery := fmt.Sprintf(`INSERT INTO "SMS" ("Device","Tittle","Data") VALUES ("%s","%s","%s");`, DeviceName, strings.Replace(Tittle, "\"", "", -1), strings.Replace(Data, "\"", "", -1))
	db.Exec(SQLQuery)
	db.Close()
}
