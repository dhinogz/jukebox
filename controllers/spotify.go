package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	llctx "github.com/dhinogz/my-tunes/context"
	"github.com/dhinogz/my-tunes/models"
	"github.com/gorilla/csrf"
)

type Spotify struct {
	Templates struct {
		Current Template
	}
	SpotifyService *models.SpotifyService
	OAuthService   *models.OAuthService
}

// func (s *Spotify) CurrentUser(w http.ResponseWriter, r *http.Request) {
// 	var data struct {
// 		U     *spotify.PrivateUser
// 		P     *spotify.SimplePlaylistPage
// 		Image string
// 	}
// 	user := llctx.User(r.Context())
// 	spToken, err := s.OAuthService.Find(user.ID, "spotify")
// 	if err != nil {
// 		if errors.Is(err, models.ErrNotFound) {
// 			http.Error(w, "Not found", http.StatusNotFound)
// 			// errors.Public(err, "Please sign in again")
// 			return
// 		}
// 		http.Error(w, "Internal server error", http.StatusInternalServerError)
// 		return
// 	}

// 	spUser, playlists, err := s.SpotifyService.Current(spToken.Token, r.Context())
// 	if err != nil {
// 		http.Error(w, "Failed to fetch Spotify user", http.StatusBadRequest)
// 		return
// 	}

// 	data.U = spUser
// 	data.P = playlists
// 	data.Image = spUser.Images[0].URL

// 	s.Templates.Current.Execute(w, r, data)
// }

func (s *Spotify) Connect(w http.ResponseWriter, r *http.Request) {
	state := csrf.Token(r)
	cookie := newCookie(CookieOAuth, state)
	http.SetCookie(w, cookie)
	url := s.SpotifyService.OAuth2.AuthCodeURL(state)
	fmt.Println(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Spotify) Callback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	state := r.FormValue("state")
	cookie, err := r.Cookie(CookieOAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if cookie == nil || cookie.Value != state {
		http.Error(w, "Invalid state provided", http.StatusBadRequest)
		return
	}
	cookie.Value = ""
	cookie.Expires = time.Now()
	http.SetCookie(w, cookie)

	code := r.FormValue("code")
	token, err := s.SpotifyService.OAuth2.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := llctx.User(r.Context())
	existing, err := s.OAuthService.Find(user.ID, "spotify")
	if err == models.ErrNotFound {
		// noop
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		s.OAuthService.Delete(existing.ID)
	}
	userOAuth := models.OAuth{
		UserID:   user.ID,
		Provider: "spotify",
		Token:    *token,
	}
	err = s.OAuthService.Create(&userOAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/home", http.StatusFound)
}
