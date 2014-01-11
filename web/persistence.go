package web

import (
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

// TODO: Is sql.DB thread-safe?
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
	stmt, err := conn.Prepare("delete from RegisteredUsers where address = ?")
	if err != nil {
		return false
	}
	_, err = stmt.Exec(address)
	if err != nil {
		return false
	}
	return true
}
