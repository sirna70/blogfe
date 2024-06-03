package models

import (
	"time"
)

type Post struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	PublishDate time.Time `json:"publish_date"`
}
