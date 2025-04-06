package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"

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

func DeleteService(db *sql.DB, service string) error {
	// TO IMPLEMENT
	service_exists, err := CheckServiceExists(db, service)
	if err != nil {
		return err
	}
	if !service_exists {
		return errors.New("el servicio no existe")
	}

	var query string = "DELETE FROM services WHERE name = ?"
	_, err = db.Exec(query, service)
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

func UpdatePassword(db *sql.DB, service string, newPassword string) error {
	query := "UPDATE services SET password = ? WHERE name = ?"
	_, err := db.Exec(query, newPassword, service)
	return err
}

func DeriveKey(passphrase string, salt string) []byte {
	return pbkdf2.Key([]byte(passphrase), []byte(salt), 4096, 32, sha256.New)
}

// Cifra un texto con AES-GCM
func EncryptAES(plaintext, passphrase string) (string, error) {
	salt := sha3.Sum256([]byte(passphrase))       // Generamos una "sal" fija desde la frase
	key := DeriveKey(passphrase, string(salt[:])) // Derivamos la clave

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Descifra un texto con AES-GCM
func DecryptAES(ciphertext, passphrase string) (string, error) {
	salt := sha3.Sum256([]byte(passphrase))
	key := DeriveKey(passphrase, string(salt[:]))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func CreateSecurePassword() string {
	// to implement
	return ""
}
