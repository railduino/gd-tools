package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var (
	jwt_secret []byte
)

func LoginHashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func LoginCheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func LoginCreateJWT(email string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwt_secret)
}

func LoginExtractJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return jwt_secret, nil
		})
	if err != nil {
		return "", fmt.Errorf("token konnte nicht validiert werden: %v", err)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims.Email, nil
	}

	return "", fmt.Errorf("ungültige Claims oder Token")
}

func UserLoginGet(w http.ResponseWriter, r *http.Request, email string) {
	msgWarning := "Unbekannte Email-Adresse oder falsches Passwort"

	data := struct {
		Email string
	}{
		Email: email,
	}
	content, err := ServeRender(w, r, "user_login", data)
	if err != nil {
		return
	}

	page := Page{
		Title:   WebT(r, "LoginTitle"),
		Content: content,
	}
	if email != "" {
		page.Warning = msgWarning
	}

	page.Serve(w, r)
}

func UserLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Fehler beim Parsen des Formulars", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	var user User
	if err := ServeDB.Where("email = ? AND login = ?", email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			UserLoginGet(w, r, email)
		} else {
			http.Error(w, "Fehler bei der Datenbank-Abfrage", http.StatusInternalServerError)
		}
		return
	}

	if err := LoginCheckPassword(user.Password, password); err != nil {
		UserLoginGet(w, r, email)
		return
	}

	accessToken, err := LoginCreateJWT(user.Email, 15*time.Minute)
	if err != nil {
		http.Error(w, "Fehler beim Erstellen des Access-Tokens", http.StatusInternalServerError)
		return
	}

	refreshToken, err := LoginCreateJWT(user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Fehler beim Erstellen des Refresh-Tokens", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: false,
		Secure:   os.Getenv("JWT_SECRET") != "",
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: false,
		Secure:   os.Getenv("JWT_SECRET") != "",
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	ServeSuccess(w, r, "LoginSuccess")
	http.Redirect(w, r, "/", http.StatusFound)
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	if current := UserCurrent(w, r, false); current != nil {
		current.ClearCSRFToken()
	}

	cookie := &http.Cookie{
		Name:   "refresh_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:   "access_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	ServeSuccess(w, r, "LoggedOut")
	http.Redirect(w, r, "/", http.StatusFound)
}

func UserLoginInit() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "get_secret_from_env"
	}
	jwt_secret = []byte(secret)

	ServeMux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			UserLoginPost(w, r)
		} else {
			UserLoginGet(w, r, "")
		}
	})
	ServeMux.Handle("/logout", http.HandlerFunc(UserLogout))
}
