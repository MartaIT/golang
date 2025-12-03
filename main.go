package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Mahasiswa struct {
	ID    int    `json:"id"`
	NIM   string `json:"nim"`
	Nama  string `json:"nama"`
	Prodi string `json:"prodi"`
	Email string `json:"email"`
}

var db *sql.DB

func connectDB() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/cloud_computing" // ⚠️ sesuaikan user/password MySQL
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Koneksi ke database berhasil")
}

// 🔹 GET: ambil semua data
func getAllMahasiswa(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, nim, nama, prodi, email FROM mahasiswa")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var data []Mahasiswa
	for rows.Next() {
		var m Mahasiswa
		rows.Scan(&m.ID, &m.NIM, &m.Nama, &m.Prodi, &m.Email)
		data = append(data, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// 🔹 POST: tambah data
func createMahasiswa(w http.ResponseWriter, r *http.Request) {
	var m Mahasiswa
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err := db.Exec("INSERT INTO mahasiswa (nim, nama, prodi, email) VALUES (?, ?, ?, ?)",
		m.NIM, m.Nama, m.Prodi, m.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// 🔹 PUT: update data
func updateMahasiswa(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m Mahasiswa
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err := db.Exec("UPDATE mahasiswa SET nim=?, nama=?, prodi=?, email=? WHERE id=?",
		m.NIM, m.Nama, m.Prodi, m.Email, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// 🔹 DELETE: hapus data
func deleteMahasiswa(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM mahasiswa WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	connectDB()
	r := mux.NewRouter()

	r.HandleFunc("/api/mahasiswa", getAllMahasiswa).Methods("GET")
	r.HandleFunc("/api/mahasiswa", createMahasiswa).Methods("POST")
	r.HandleFunc("/api/mahasiswa/{id}", updateMahasiswa).Methods("PUT")
	r.HandleFunc("/api/mahasiswa/{id}", deleteMahasiswa).Methods("DELETE")

	// 🔹 Serve file HTML
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./template/index.html")
	})

	fmt.Println("Server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
