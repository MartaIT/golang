package models

type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"` // hashed password
    Role     string `json:"role"` // e.g., "user" or "admin"
}
