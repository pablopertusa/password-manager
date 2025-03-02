package main

import (
	"cli-tool/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // Importación anónima para registrar el driver
)

type PageData struct {
	Title string
}

type InfoPage struct {
	Information string
}

type Password struct {
	Service string
}

func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		return nil, err
	}
	fmt.Println("Conectado a SQLite exitosamente!")

	db, e := utils.CreateTable(db)
	if e != nil {
		return nil, e
	}

	return db, nil
}

func showPasswordsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		tmpl, err := template.ParseFiles("static/passwords.html")
		if err != nil {
			http.Error(w, "Error cargando la plantilla", http.StatusInternalServerError)
			return
		}

		services, err := utils.GetServices(db)
		if err != nil {
			http.Error(w, "Error obteniendo servicios", http.StatusInternalServerError)
			return
		}

		passwords := []Password{}
		for _, service := range services {
			passwords = append(passwords, Password{Service: service})
		}

		err = tmpl.Execute(w, struct{ Passwords []Password }{Passwords: passwords})
		if err != nil {
			http.Error(w, "Error renderizando la plantilla", http.StatusInternalServerError)
		}
	}
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/home.html"))
	data := PageData{Title: "Gestor de Contraseñas"}
	tmpl.Execute(w, data)
}

func formHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/form.html"))
	tmpl.Execute(w, tmpl)
}

func addPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		service := r.FormValue("service")
		password := r.FormValue("password")
		passphrase := r.FormValue("passphrase")

		if service == "" || password == "" || passphrase == "" {
			http.Error(w, "Faltan datos", http.StatusBadRequest)
			return
		}
		encrypted_password, err := utils.EncryptAES(password, passphrase)
		if err != nil {
			tmpl := template.Must(template.ParseFiles("static/error.html"))
			pagedata := InfoPage{Information: err.Error()}
			tmpl.Execute(w, pagedata)
			return
		}
		err = utils.AddPassword(db, service, encrypted_password)
		if err != nil {
			tmpl := template.Must(template.ParseFiles("static/error.html"))
			pagedata := InfoPage{Information: err.Error()}
			tmpl.Execute(w, pagedata)
			return
		}
		tmpl := template.Must(template.ParseFiles("static/success.html"))
		info := InfoPage{Information: "Contraseña creada con éxito"}
		tmpl.Execute(w, info)
	}
}

func getPasswordPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		service := r.URL.Query().Get("service")
		exists, err := utils.CheckServiceExists(db, service)
		if service == "" {
			http.Error(w, "Parámetro 'service' es requerido", http.StatusBadRequest)
			return
		}
		if err != nil {
			tmpl := template.Must(template.ParseFiles("static/error.html"))
			pagedata := InfoPage{Information: err.Error()}
			tmpl.Execute(w, pagedata)
			return
		}
		if !exists {
			tmpl := template.Must(template.ParseFiles("static/error.html"))
			pagedata := InfoPage{Information: "El servicio no existe"}
			tmpl.Execute(w, pagedata)
			return
		}

		tmpl := template.Must(template.ParseFiles("static/decrypt.html"))
		info := InfoPage{Information: service}
		tmpl.Execute(w, info)
	}
}

func decryptPasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("service")
		passphrase := r.URL.Query().Get("passphrase")
		ciphertext, err := utils.GetPassword(db, service)
		if err != nil {
			response := map[string]string{"error": err.Error()}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		plaintext, err := utils.DecryptAES(ciphertext, passphrase)
		if err != nil {
			response := map[string]string{"error": "Error al desencriptar, comprueba tu passphrase"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		response := map[string]string{"password": plaintext}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func getIdentityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "peibol" {
		http.Redirect(w, r, "/home", http.StatusSeeOther) // 303 See Other para redirecciones después de POST
	} else {
		w.Write([]byte("vete de aquí"))
	}
}

func identityFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/identity.html"))
	tmpl.Execute(w, tmpl)
}

func main() {

	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", identityFormHandler)
	http.HandleFunc("/get-identity", getIdentityHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/show-passwords", showPasswordsHandler(db))
	http.HandleFunc("/add-password", addPasswordHandler(db))
	http.HandleFunc("/add-password-form", formHandler)
	http.HandleFunc("/get-password", getPasswordPageHandler(db))
	http.HandleFunc("/decrypt-password", decryptPasswordHandler(db))
	fmt.Println("init")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":2727", nil)
}
