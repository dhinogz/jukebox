package models

import (
	"context"
	"database/sql"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyCfg struct {
	OAuth OAuthSettings
}

type SpotifyService struct {
	DB     *sql.DB
	OAuth2 *oauth2.Config
}

type SpotifyUser struct {
	spotify.PrivateUser
}

func NewSpotifyService(db *sql.DB, cfg SpotifyCfg) *SpotifyService {
	oauth2 := buildConfig(&cfg.OAuth)

	ds := &SpotifyService{
		DB:     db,
		OAuth2: oauth2,
	}

	return ds
}

func createSpotifyClient(token oauth2.Token, ctx context.Context) *spotify.Client {
	httpClient := spotifyauth.New().Client(ctx, &token)
	return spotify.New(httpClient)
}

func currentUser(spToken oauth2.Token, ctx context.Context) (*spotify.PrivateUser, error) {
	client := createSpotifyClient(spToken, ctx)
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func currentUserPlaylists(spToken oauth2.Token, ctx context.Context) (*spotify.SimplePlaylistPage, error) {
	client := createSpotifyClient(spToken, ctx)
	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}
