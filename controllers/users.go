package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dhinogz/my-tunes/context"
	"github.com/dhinogz/my-tunes/errors"
	"github.com/dhinogz/my-tunes/models"
	"github.com/zmb3/spotify/v2"
)

type Users struct {
	Templates struct {
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
		Current        Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
	OAuthService         *models.OAuthService
	PostService          *models.PostService
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	var data struct {
		U     *spotify.PrivateUser
		P     *spotify.SimplePlaylistPage
		Image string
		Posts []models.Post
	}
	user := context.User(r.Context())
	spToken, err := u.OAuthService.Find(user.ID, "spotify")
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Not found", http.StatusNotFound)
			// errors.Public(err, "Please sign in again")
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	spUser, err := u.UserService.Current(spToken.Token, r.Context())
	playlists, err := u.UserService.CurrentUserPlaylist(spToken.Token, r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch Spotify user", http.StatusBadRequest)
		return
	}

	posts, err := u.PostService.ByAuthorID(user.ID)
	if err != nil {
		http.Error(w, "Not Found", http.StatusInternalServerError)
		return
	}

	data.U = spUser
	data.P = playlists
	data.Image = spUser.Images[0].URL
	data.Posts = posts

	u.Templates.Current.Execute(w, r, data)
}

// New user sign up template
func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

// Process new user sign up
func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Create(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken) {
			err = errors.Public(err, "That email address is already in use.")
		}
		u.Templates.New.Execute(w, r, data, err)
		return
	}
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		// TODO: long term a warning should be shown stating failed sign in
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/home", http.StatusFound)
}

// Sign in user template
func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

// Process a user's sign in
func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrAuth) {
			err = errors.Public(err, "Could not find a user with that email or password.")
		}
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/home", http.StatusFound)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.Delete(token)
	if err != nil {
		if errors.Is(err, models.ErrSignOut) {
			err = errors.Public(err, "Could not sign you out.")
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	deleteCookie(w, CookieOAuth)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		if errors.Is(err, models.ErrEmailDoesNotExist) {
			err = errors.Public(err, "Could not find a user with that email or password.")
		}
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}
	vals := url.Values{
		"token": {pwReset.Token},
	}
	// TODO: make the url here configurable
	err = u.EmailService.ForgotPassword(data.Email, "https://www.mytunes.com/reset-pw?"+vals.Encode())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		// TODO: Distinguish between server errors and invalid token errors.
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// Sign the user in now that they have reset their password.
	// Any errors from this point onwards should redirect to the sign in page
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

type UserMiddleware struct {
	SessionService *models.SessionService
	OAuthService   *models.OAuthService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		ctx = context.WithUser(ctx, user)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireOAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

		spUser, _ := umw.OAuthService.Find(user.ID, "spotify")
		if spUser == nil {
			http.Redirect(w, r, "/oauth/spotify/connect", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
