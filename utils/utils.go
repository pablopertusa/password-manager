package utils

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

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
		return errors.New("el servicio ya existe, edítalo si quieres cambiar la contraseña")
	}

	var query string = "INSERT INTO services (name, password) VALUES (?, ?)"
	_, err = db.Exec(query, service, password)
	if err != nil {
		return err
	}
	return nil
}

func GetPassword(db *sql.DB, service string) (string, error) {
	exists, err := CheckServiceExists(db, service)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("no hay ningún servicio con ese nombre")
	}
	var password string
	query := "SELECT password FROM services WHERE name = ?"
	err = db.QueryRow(query, service).Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}
