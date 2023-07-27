package controllers

import (
	"net/http"

	"github.com/dhinogz/my-tunes/context"
	"github.com/dhinogz/my-tunes/models"
)

type Posts struct {
	Templates struct {
		CurrentUser Template
		Feed        Template
	}
	PostService *models.PostService
}

func (p Posts) CurrentUser(w http.ResponseWriter, r *http.Request) {
	type Post struct {
		ID             int
		AuthorID       int
		Title          string
		Content        string
		SpotifyContext string
	}
	var data struct {
		Posts []Post
	}

	user := context.User(r.Context())
	posts, err := p.PostService.ByAuthorID(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, post := range posts {
		data.Posts = append(data.Posts, Post{
			ID:             post.ID,
			AuthorID:       post.AuthorID,
			Title:          post.Title,
			Content:        post.Content,
			SpotifyContext: post.SpotifyContext,
		})
	}

	p.Templates.Feed.Execute(w, r, data)
}
