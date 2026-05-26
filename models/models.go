package models

import "time"

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	UUID      string
	UserID    int
	ExpiresAt time.Time
}

type Category struct {
	ID   int
	Name string
	Slug string
}

type Topic struct {
	ID         int
	UserID     int
	CategoryID int
	Title      string
	CreatedAt  time.Time
	// champs enrichis (JOIN)
	Username     string
	CategoryName string
	CategorySlug string
	PostCount    int
	LikeCount    int
}

type Post struct {
	ID        int
	TopicID   int
	UserID    int
	Content   string
	ImageURL  string
	CreatedAt time.Time
	// champs enrichis (JOIN)
	Username     string
	Likes        int
	Dislikes     int
	UserReaction string // "like", "dislike", ou ""
}

type Reaction struct {
	UserID int
	PostID int
	Type   string // "like" ou "dislike"
}

type PageData struct {
	User       *User
	Categories []Category
	Topics     []Topic
	Topic      *Topic
	Category   *Category
	Posts      []Post
	Error      string
	Filter     string
	Search     string
}
