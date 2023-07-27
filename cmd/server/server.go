package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/dhinogz/my-tunes/controllers"
	"github.com/dhinogz/my-tunes/migrations"
	"github.com/dhinogz/my-tunes/models"
	"github.com/dhinogz/my-tunes/templates"
	"github.com/dhinogz/my-tunes/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL    models.PostgresConfig
	SMTP    models.SMTPConfig
	Spotify models.SpotifyCfg
	CSRF    struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	// Postgres config
	cfg.PSQL = models.PostgresConfig{
		Host:        os.Getenv("PSQL_HOST"),
		Port:        os.Getenv("PSQL_PORT"),
		User:        os.Getenv("PSQL_USER"),
		Password:    os.Getenv("PSQL_PASSWORD"),
		Database:    os.Getenv("PSQL_DATABASE"),
		SSLMode:     os.Getenv("PSQL_SSLMODE"),
		MaxIdleTime: os.Getenv("MAX_IDLE_TIME"),
	}
	maxIdleConnsStr := os.Getenv("MAX_IDLE_CONNS")
	cfg.PSQL.MaxIdleConns, err = strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		return cfg, err
	}
	maxOpenConnsStr := os.Getenv("MAX_OPEN_CONNS")
	cfg.PSQL.MaxOpenConns, err = strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		return cfg, err
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("No PSQL Config provided.")
	}

	// SMTP config
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// Spotify config
	cfg.Spotify.OAuth.ClientID = os.Getenv("SPOTIFY_CLIENT_ID")
	cfg.Spotify.OAuth.ClientSecret = os.Getenv("SPOTIFY_CLIENT_SECRET")
	cfg.Spotify.OAuth.AuthURL = os.Getenv("SPOTIFY_AUTH_URL")
	cfg.Spotify.OAuth.TokenURL = os.Getenv("SPOTIFY_TOKEN_URL")
	cfg.Spotify.OAuth.RedirectURL = os.Getenv("SPOTIFY_REDIRECT_URL")
	cfg.Spotify.OAuth.Scopes = os.Getenv("SPOTIFY_SCOPES")

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"

	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")

	return cfg, nil
}

func main() {
	// Set up config
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}
	err = run(cfg)
	if err != nil {
		panic(err)
	}

}

func run(cfg config) error {
	// Set up database
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		return err
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		return err
	}

	// Set up services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	oauthService := &models.OAuthService{
		DB: db,
	}
	emailService, err := models.NewEmailService(cfg.SMTP)
	if err != nil {
		return err
	}
	spotifyService := models.NewSpotifyService(db, cfg.Spotify)
	postService := &models.PostService{
		DB: db,
	}

	// Set up middleware
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
		OAuthService:   oauthService,
	}

	// Set up controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
		OAuthService:         oauthService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.html", "tailwind.html",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.html", "tailwind.html",
	))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"forgot-pw.html", "tailwind.html",
	))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.html", "tailwind.html",
	))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"reset-pw.html", "tailwind.html",
	))
	usersC.Templates.Current = views.Must(views.ParseFS(
		templates.FS,
		"current-user.html", "tailwind.html",
	))

	spotifyC := controllers.Spotify{
		SpotifyService: spotifyService,
		OAuthService:   oauthService,
	}

	postsC := controllers.Posts{
		PostService: postService,
	}

	postsC.Templates.Feed = views.Must(views.ParseFS(
		templates.FS,
		"feed.html", "tailwind.html",
	))

	// galleriesC := controllers.Galleries{
	// 	GalleryService: galleryService,
	// }
	// galleriesC.Templates.New = views.Must(views.ParseFS(
	// 	templates.FS,
	// 	"galleries/new.html", "tailwind.html",
	// ))
	// galleriesC.Templates.Edit = views.Must(views.ParseFS(
	// 	templates.FS,
	// 	"galleries/edit.html", "tailwind.html",
	// ))
	// galleriesC.Templates.Index = views.Must(views.ParseFS(
	// 	templates.FS,
	// 	"galleries/index.html", "tailwind.html",
	// ))
	// galleriesC.Templates.Show = views.Must(views.ParseFS(
	// 	templates.FS,
	// 	"galleries/show.html", "tailwind.html",
	// ))
	// Set up router
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(csrfMw)
	r.Use(umw.SetUser)

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"home.html", "tailwind.html",
	))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"contact.html", "tailwind.html",
	))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(
		templates.FS,
		"faq.html", "tailwind.html",
	))))
	r.Get("/check-email", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.html", "tailwind.html",
	))))
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Route("/signup", func(r chi.Router) {
		r.Get("/", usersC.New)
		r.Post("/", usersC.Create)
	})

	r.Route("/signin", func(r chi.Router) {
		r.Get("/", usersC.SignIn)
		r.Post("/", usersC.ProcessSignIn)
	})

	r.Post("/signout", usersC.ProcessSignOut)

	r.Route("/oauth/spotify", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/connect", spotifyC.Connect)
		r.Get("/callback", spotifyC.Callback)
	})

	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireOAuth)
		r.Get("/", usersC.CurrentUser)
	})

	r.Route("/home", func(r chi.Router) {
		r.Use(umw.RequireOAuth)
		r.Get("/", postsC.Feed)
	})

	// r.Route("/galleries", func(r chi.Router) {
	// 	r.Get("/{id}", galleriesC.Show)
	// 	r.Get("/{id}/images/{filename}", galleriesC.Image)
	// 	r.Group(func(r chi.Router) {
	// 		r.Use(umw.RequireUser)
	// 		r.Get("/", galleriesC.Index)
	// 		r.Get("/new", galleriesC.New)
	// 		r.Post("/", galleriesC.Create)
	// 		r.Get("/{id}/edit", galleriesC.Edit)
	// 		r.Post("/{id}", galleriesC.Update)
	// 		r.Post("/{id}/delete", galleriesC.Delete)
	// 		r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
	// 		r.Post("/{id}/images", galleriesC.UploadImage)
	// 	})
	// })

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	assetsHandler := http.FileServer(http.Dir("assets"))
	r.Get("/assets/*", http.StripPrefix("/assets", assetsHandler).ServeHTTP)

	fmt.Printf("Starting server on %s...\n", cfg.Server.Address)
	return http.ListenAndServe(":3000", r)
}
