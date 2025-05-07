package main

import (
	_ "bytes"
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Email   string `gorm:"unique;index;not null" json:"email"`
	Name    string `gorm:"unique;not null" json:"name"`
	Phone   string `json:"phone"`
	Service string `json:"service"`

	Refresh bool `gorm:"-" json:"-"`

	Enabled  bool   `json:"enabled"`
	SysAdmin bool   `json:"sys_admin"`
	Password string `json:"password"`
	Profile  bool   `gorm:"-" json:"-"`

	Token  string    `gorm:"index" json:"-"`
	Expire time.Time `json:"-"`

	// TODO Access []Access `gorm:"foreignKey:UserID"`
}

const userContextKey contextKey = "user"

var (
	CSRFTokenMap = make(map[string]string)
)

func (current User) AddCSRFToken() string {
	token, _ := NewToken(16)

	CSRFTokenMap[token] = current.Email

	return token
}

func (current User) CheckCSRFToken(token string) bool {
	if _, ok := CSRFTokenMap[token]; ok {
		delete(CSRFTokenMap, token)
		return true
	}

	return false
}

func (current User) ClearCSRFToken() {
	for key, value := range CSRFTokenMap {
		if value == current.Email {
			delete(CSRFTokenMap, key)
		}
	}
}

func (user User) EmailCRC() string {
	return CRC32(user.Email)
}

func (user User) NameCRC() string {
	return CRC32(user.Name)
}

func InitServeUser() error {
	if err := ServeDB.AutoMigrate(&User{}); err != nil {
		return err
	}

	return nil
}
