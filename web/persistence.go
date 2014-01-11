package web

import (
	"log"
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

func InitializeDB() (*DbConnection, error) {
	conn, err := sql.Open("sqlite3", "./test.db")
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
	rows, err := conn.Query("select * from RegisteredUsers where address = ?", address)
	if err != nil {
		return false
	}
	defer rows.Close()
	if rows.Next() {
		return doPreparedStatement(conn,
			"update RegisteredUsers set name=?, email=?, note=? where address=?",
			record.Name, record.Email, record.Note, address)
	} else {
		return doPreparedStatement(conn,
			"insert into RegisteredUsers values (?, ?, ?, ?)",
			address, record.Name, record.Email, record.Note)
	}
}

func doPreparedStatement(conn *DbConnection, statement string,
	arguments ...interface{}) bool {
	
	stmt, err := conn.Prepare(statement)
	if err != nil {
		log.Printf("Failed create statement %q", err.Error())
		return false
	}
	_, err = stmt.Exec(arguments...)
	if err != nil {
		log.Printf("Failed execute statement %q", err.Error())
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
