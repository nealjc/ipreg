package web

import (
	"log"
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type RegistrationRecord struct {
	Name string
	Email string
	Note string
}

type DbConnection struct {
	*sql.DB
}

func InitializeDB(dbName string) (*DbConnection, error) {
	log.Print("Opening database")
	conn, err := sql.Open("sqlite3",
		fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbName))
	if err != nil {
		return nil, err
	}
	
	sql := `
create table if not exists RegisteredUsers (address text primary key, name text,
email text, note text);
`
	_, err = conn.Exec(sql)
	if err != nil {
		return nil, err
	}

	return &DbConnection{conn}, nil
}

func (conn *DbConnection) SetRegistration(address string,
	record RegistrationRecord) bool {
	// TODO: need a lock here
	rows, err := conn.Query("select * from RegisteredUsers where address = ?;", address)
	if err != nil {
		return false
	}
	if rows.Next() {
		rows.Close()
		return doPreparedStatement(conn,
			"update RegisteredUsers set name=?, email=?, note=? where address=?;",
			record.Name, record.Email, record.Note, address)
	} else {
		rows.Close()
		return doPreparedStatement(conn,
			"insert into RegisteredUsers values (?, ?, ?, ?);",
			address, record.Name, record.Email, record.Note)
	}
}

func doPreparedStatement(conn *DbConnection, statement string,
	arguments ...interface{}) bool {
	
	result, err := conn.Exec(statement, arguments...)
	if err != nil {
		log.Printf("Failed execute statement %q", err.Error())
		return false
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return false
	} 
	return true
}

func (conn *DbConnection) GetRegistration(address string) (RegistrationRecord, error) {
	rows, err := conn.Query("select * from RegisteredUsers where address = ?", address)
	if err != nil {
		return RegistrationRecord{"", "", ""}, err
	}
	defer rows.Close()
	if rows.Next() {
		var address, name, email, note string
		rows.Scan(&address, &name, &email, &note)
		return RegistrationRecord{name, email, note}, nil
	}
	return RegistrationRecord{"", "", ""}, nil
}

func (conn *DbConnection)  DeleteRegistration(address string) bool {
	return doPreparedStatement(conn, "delete from RegisteredUsers where address = ?", address)
}
