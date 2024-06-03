package models

type Tag struct {
	ID      int    `json:"id"`
	Label   string `json:"label" validate:"required"`
	PostsID int64  `json:"posts_id"`
}
