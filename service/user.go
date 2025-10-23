package service

import (
	"net/http"

	"backgammon/business"
	"backgammon/util"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	token := business.LogInUser()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		// Secure: true, // enable when using HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	util.JSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

}

func RegisterTokenHandler(w http.ResponseWriter, r *http.Request) {

}

func SessionHandler(w http.ResponseWriter, r *http.Request) {

}
