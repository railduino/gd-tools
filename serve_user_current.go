package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

func (user User) LoginOptions() template.HTML {
	var optList []string

	if user.Enabled {
		optList = append(optList, "<option value='off'>Gesperrt</option>")
		optList = append(optList, "<option value='on' selected>Erlaubt</option>")
	} else {
		optList = append(optList, "<option value='off' selected>Gesperrt</option>")
		optList = append(optList, "<option value='on'>Erlaubt</option>")
	}

	return template.HTML(strings.Join(optList, "\n"))
}

func UserCurrent(w http.ResponseWriter, r *http.Request, sysAdmin bool) *User {
	var current User

	cookie, err := r.Cookie("access_token")
	if err == nil {
		if email, err := LoginExtractJWT(cookie.Value); err == nil {
			if err := ServeDB.Where("email = ?", email).First(&current).Error; err == nil {
				log.Printf("INFO: '%s' found via AccessToken", current.Email)
				if sysAdmin && !current.SysAdmin {
					Unauthorized(w, r, "/")
					return nil
				}
				current.Refresh = false
				return &current
			}
		}
	}

	cookie, err = r.Cookie("refresh_token")
	if err == nil {
		if email, err := LoginExtractJWT(cookie.Value); err == nil {
			if err := ServeDB.Where("email = ?", email).First(&current).Error; err == nil {
				log.Printf("INFO: '%s' found via RefreshToken", current.Email)
				if sysAdmin && !current.SysAdmin {
					Unauthorized(w, r, "/")
					return nil
				}
				current.Refresh = true
				return &current
			}
		}
	}

	ServeWarning(w, r, "PleaseLogin")
	http.Redirect(w, r, "/login", http.StatusFound)

	return nil
}

func Unauthorized(w http.ResponseWriter, r *http.Request, target string) {
	ServeWarning(w, r, "NotAllowed")
	http.Redirect(w, r, target, http.StatusFound)
}
