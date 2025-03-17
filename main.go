package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"password-manager/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// Cargar variables de entorno
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("Error cargando el archivo .env")
	}
}

func init() {
	loadEnv()
}

var rootName string
var jwtKey []byte

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
	db, err := sql.Open("sqlite3", "database.db")
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

func identityFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/identity.html"))
	tmpl.Execute(w, nil)
}

// Middleware para verificar JWT
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Acceso denegado: no autenticado", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Token mal formado", http.StatusUnauthorized)
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				http.Error(w, "Token expirado", http.StatusUnauthorized)
				return
			}
		}

		username, _ := claims["username"].(string)
		fmt.Println("Usuario autenticado:", username, ", timestamp:", time.Now().String())

		next.ServeHTTP(w, r) // Si el token es válido, continúa con la solicitud
	})
}

// Generar JWT para el usuario
func getIdentityHandler(rootName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		name := r.FormValue("name")
		if name == rootName {
			expirationTime := time.Now().Add(15 * time.Minute)
			claims := jwt.MapClaims{
				"username": name,
				"exp":      expirationTime.Unix(),
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(jwtKey)
			if err != nil {
				http.Error(w, "Error generando token", http.StatusInternalServerError)
				return
			}

			// Configurar la cookie con el token
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    tokenString,
				Expires:  expirationTime,
				HttpOnly: true,  // Protege contra acceso desde JS
				Secure:   false, // Ponlo en true si usas HTTPS
				Path:     "/",
			})

			http.Redirect(w, r, "/protected/home", http.StatusSeeOther)
		} else {
			w.Write([]byte("vete de aquí"))
		}
	}
}

func updatePasswordHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := r.URL.Query().Get("service")
		newPassword := r.URL.Query().Get("newPassword")
		passphrase := r.URL.Query().Get("passphrase")

		if service == "" || newPassword == "" || passphrase == "" {
			http.Error(w, "Faltan datos", http.StatusBadRequest)
			return
		}

		encrypted_password, err := utils.EncryptAES(newPassword, passphrase)
		if err != nil {
			tmpl := template.Must(template.ParseFiles("static/error.html"))
			pagedata := InfoPage{Information: err.Error()}
			tmpl.Execute(w, pagedata)
			return
		}

		err = utils.UpdatePassword(db, service, encrypted_password)
		if err != nil {
			response := map[string]bool{"success": false}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		response := map[string]bool{"success": true}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func main() {

	rootName = os.Getenv("USER_NAME")
	jwtKey = []byte(os.Getenv("JWT_KEY"))

	if rootName == "" {
		log.Fatal("Error: 'USER_NAME' no definido en .env")
	}
	if len(jwtKey) == 0 {
		log.Fatal("Error: 'JWT_KEY' no definido en .env")
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()

	r.HandleFunc("/", identityFormHandler).Methods("GET")
	r.HandleFunc("/login", getIdentityHandler(rootName)).Methods("POST")

	// Todas las rutas protegidas pasan por el middleware
	protected := r.PathPrefix("/protected").Subrouter()
	protected.Use(authMiddleware)
	protected.HandleFunc("/home", homeHandler).Methods("GET")
	protected.HandleFunc("/show-passwords", showPasswordsHandler(db)).Methods("GET")
	protected.HandleFunc("/add-password", addPasswordHandler(db)).Methods("POST")
	protected.HandleFunc("/add-password-form", formHandler).Methods("GET")
	protected.HandleFunc("/get-password", getPasswordPageHandler(db)).Methods("GET")
	protected.HandleFunc("/decrypt-password", decryptPasswordHandler(db)).Methods("GET")
	protected.HandleFunc("/update-password", updatePasswordHandler(db)).Methods("POST")
	protected.PathPrefix("/static/").Handler(http.StripPrefix("/protected/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Servidor en http://localhost:2727")

	http.ListenAndServe(":2727", r)
}
