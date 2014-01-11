package web

type RegistrationRecord struct {
	Name string
	Email string
	Note string
}

func InitializeDB() bool {
	return true
}

func SetRegistration(address string, record RegistrationRecord) bool {
	return true
}

func GetRegistration(address string) (RegistrationRecord, error) {
	return RegistrationRecord{"dummy", "dummy", "dummy"}, nil
}

func DeleteRegistration(address string) bool {
	return true
}
