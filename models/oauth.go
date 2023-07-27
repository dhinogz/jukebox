package models

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dhinogz/my-tunes/errors"
	"golang.org/x/oauth2"
)

type OAuthSettings struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
	Scopes       string
}

type OAuth struct {
	ID       int
	UserID   int    // not null, unique index
	Provider string // not null
	Token    oauth2.Token
}

type OAuthService struct {
	DB *sql.DB
}

func (os *OAuthService) Find(userID int, provider string) (*OAuth, error) {
	var oauth OAuth
	stmt := `SELECT id, user_id, provider, access_token, refresh_token, token_type, expiry 
		FROM oauth 
		WHERE user_id = $1 AND provider = $2;`
	row := os.DB.QueryRow(stmt, userID, provider)
	err := row.Scan(&oauth.ID, &oauth.UserID, &oauth.Provider, &oauth.Token.AccessToken, &oauth.Token.RefreshToken, &oauth.Token.TokenType, &oauth.Token.Expiry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find oauth: %w", err)
	}
	return &oauth, nil
}

func (os *OAuthService) Create(userIn *OAuth) error {
	var oauthID int
	stmt := `INSERT INTO oauth (user_id, provider, access_token, refresh_token, token_type, expiry)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ID;`
	row := os.DB.QueryRow(stmt, userIn.UserID, userIn.Provider, userIn.Token.AccessToken, userIn.Token.RefreshToken, userIn.Token.TokenType, userIn.Token.Expiry)
	err := row.Scan(&oauthID)
	if err != nil {
		return fmt.Errorf("create oauth: %w", err)
	}
	return nil
}

func (os *OAuthService) Delete(userID int) {
	stmt := `DELETE FROM oauth
		WHERE user_id = $1`
	os.DB.Exec(stmt, userID)
}

func buildConfig(cfg *OAuthSettings) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  cfg.AuthURL,
		TokenURL: cfg.TokenURL,
	}
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     endpoint,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       strings.Split(cfg.Scopes, ","),
	}
}
