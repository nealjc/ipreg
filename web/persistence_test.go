package web

import (
	"testing"
	"os"
)

var testDbFile string = "./tmp.db"

func createTestDbFile(t *testing.T)  *DbConnection {
	err := os.Remove(testDbFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal("Failed to remove test db file")		
	}
	conn, err := InitializeDB(testDbFile)
	if err != nil {
		t.Fatal("Failed to initialize the DB")
	}
	return conn
}

func TestNewDatabase(t *testing.T) {
	conn := createTestDbFile(t)
	defer conn.Close()
	if _, err := os.Stat(testDbFile); os.IsNotExist(err) {
		t.Fatal("DB was not created")
	}
}

func TestInsertNew(t *testing.T) {
	conn := createTestDbFile(t)
	defer conn.Close()

	insertedRecord := RegistrationRecord{"Name", "Email", "Note"}
	callSet("192.168.1.1", insertedRecord, conn, t)
}

func callSet(addr string, rec RegistrationRecord, conn *DbConnection, t *testing.T) {
	if !conn.SetRegistration(addr, rec) {
		t.Fatal("Failed to insert new record")
	}
	record, err := conn.GetRegistration(addr)
	if err != nil {
		t.Fatal("Error reteiving inserted record")
	}
	if record != rec {
		t.Fatal("Retrieved record did not match inserted")
	}
}

func TestUpdateExisting(t *testing.T) {
	conn := createTestDbFile(t)
	defer conn.Close()

	insertedRecord := RegistrationRecord{"Name", "Email", "Note"}
	callSet("192.168.1.1", insertedRecord, conn, t)
	insertedRecord.Name = "UpdatedName"
	callSet("192.168.1.1", insertedRecord, conn, t)
}

func TestDeleteExisting(t *testing.T) {
	conn := createTestDbFile(t)
	defer conn.Close()

	insertedRecord := RegistrationRecord{"Name", "Email", "Note"}
	emptyRecord := RegistrationRecord{"", "", ""}
	callSet("192.168.1.1", insertedRecord, conn, t)
	if !conn.DeleteRegistration("192.168.1.1") {
		t.Fatal("Failed to delete")
	}
	record, err := conn.GetRegistration("192.168.1.1")
	if err != nil {
		t.Fatal("Failed to check for deleted entry")
	}
	if record != emptyRecord {
		t.Fatal("Failed to delete")
	}
}

func TestDeleteInvalid(t *testing.T) {
	conn := createTestDbFile(t)
	defer conn.Close()

	if conn.DeleteRegistration("192.168.1.1") {
		t.Fatal("Delete of invalid entry did not return error")
	}
}

func TestConcurrentAccess(t *testing.T) {
}






