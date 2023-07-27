package models

import (
	"database/sql"
	"fmt"
)

type Post struct {
	ID             int
	AuthorID       int
	Title          string
	Content        string
	SpotifyContext string
}

type PostService struct {
	DB *sql.DB
}

func (ps *PostService) ByAuthorID(authorID int) ([]Post, error) {
	stmt := `SELECT id, title, content, spotify_context
	FROM posts
	WHERE author_id = $1`
	rows, err := ps.DB.Query(stmt, authorID)
	if err != nil {
		return nil, fmt.Errorf("query posts by author: %w", err)
	}
	var posts []Post
	for rows.Next() {
		post := Post{
			AuthorID: authorID,
		}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.SpotifyContext)
		if err != nil {
			return nil, fmt.Errorf("query posts by author: %w", err)
		}
		posts = append(posts, post)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("query by author: %w", err)
	}
	return posts, nil
}
