package utils

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type CustomError struct {
	Message string
}

func (ce *CustomError) Error() string {
	return ce.Message
}

func CreateTable(db *sql.DB) (*sql.DB, error) {
	query := `
	CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		password TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		return db, err
	}
	return db, nil
}

func GetServices(db *sql.DB) ([]string, error) {
	var resul []string = []string{}
	rows, err := db.Query("SELECT name FROM services")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		resul = append(resul, name)
	}
	return resul, nil
}

func CheckServiceExists(db *sql.DB, service string) (bool, error) {
	query := "SELECT COUNT(*) FROM services WHERE name = ?"
	var count int

	err := db.QueryRow(query, service).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func AddPassword(db *sql.DB, service string, password string) error {
	service_exists, err := CheckServiceExists(db, service)
	if err != nil {
		return err
	}
	if service_exists {
		return &CustomError{Message: "El servicio ya existe"}
	}

	var query string = "INSERT INTO services (name, password) VALUES (?, ?)"
	_, err = db.Exec(query, service, password)
	if err != nil {
		return err
	}
	return nil
}
