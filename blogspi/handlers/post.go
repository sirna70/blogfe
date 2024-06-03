package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"blogspi/handlers/middleware"
	"blogspi/models"
	"blogspi/utils"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}
	if claims.Role != "user" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		fmt.Println("ERROR JSON ==", err)
		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	post.Status = "draft"
	post.PublishDate = time.Time{}

	tags := make([]models.Tag, len(post.Tags))
	for i, val := range post.Tags {
		tags[i] = models.Tag{Label: val}
	}

	db := utils.ConnectDB()
	defer db.Close()

	err = db.QueryRow("INSERT INTO posts (title, content, status) VALUES ($1, $2, $3) RETURNING id",
		post.Title, post.Content, post.Status).Scan(&post.ID)
	if err != nil {
		fmt.Println("Error", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	for _, tag := range post.Tags {
		_, err := db.Exec("INSERT INTO tags (label, posts_id) VALUES ($1, $2)", tag, post.ID)
		if err != nil {
			fmt.Println("Error inserting tag:", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)

}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	if claims.Role != "user" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	var post models.Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {

		http.Error(w, "Bad Request: Invalid JSON", http.StatusBadRequest)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var status string
	err = db.QueryRow("SELECT status FROM posts WHERE id=$1", post.ID).Scan(&status)
	if err != nil {
		fmt.Println("Error getting post status:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	if status != "draft" {
		http.Error(w, "Forbidden: the status posts is not 'draft', only status 'draft' can be updated", http.StatusForbidden)
		return
	}

	_, err = db.Exec("UPDATE posts SET title=$1, content=$2 WHERE id=$3",
		post.Title, post.Content, post.ID)
	if err != nil {
		fmt.Println("Error updating post:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM tags WHERE posts_id=$1", post.ID)
	if err != nil {
		fmt.Println("Error deleting tags:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	for _, tag := range post.Tags {
		_, err := db.Exec("INSERT INTO tags (label, posts_id) VALUES ($1, $2)", tag, post.ID)
		if err != nil {
			fmt.Println("Error inserting tag:", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Post successfully updated")

}

func PublishPost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	id := r.URL.Query().Get("id")
	db := utils.ConnectDB()
	defer db.Close()

	_, err := db.Exec("UPDATE posts SET status='publish', publish_date=$1 WHERE id=$2", time.Now(), id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jsonResponse := map[string]string{"message": "Admin successfully to publish", "status": "success to publish"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonResponse)
}

func DeletePost(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	db := utils.ConnectDB()
	defer db.Close()

	var status string
	err := db.QueryRow("SELECT status FROM posts WHERE id = $1", id).Scan(&status)
	if err != nil {
		fmt.Println("Error getting post status:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	if status == "publish" && claims.Role != "admin" {
		http.Error(w, "Forbidden: Only admin can delete published posts", http.StatusForbidden)
		return
	}

	if status == "draft" && claims.Role != "admin" && claims.Role != "user" {
		http.Error(w, "Forbidden: Invalid Role", http.StatusForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM tags WHERE posts_id = $1", id)
	if err != nil {
		fmt.Println("Error deleting tags:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM posts WHERE id = $1", id)
	if err != nil {
		fmt.Println("Error deleting post:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}

	jsonResponse := map[string]string{"message": "Post successfully deleted", "status": "success"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonResponse)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var rows *sql.Rows
	var err error

	if claims.Role == "admin" || claims.Role == "user" {
		rows, err = db.Query(`SELECT p.id, p.title, p.content, p.status, p.publish_date, ARRAY_AGG(t.label) AS tags
		FROM posts p JOIN tags t ON p.id = t.posts_id GROUP BY p.id, p.title, p.content, p.status, p.publish_date`)
		if err != nil {
			fmt.Println("MASUKKK==", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}
	}

	defer rows.Close()

	var postsMap = make(map[int]*models.Post)

	for rows.Next() {
		var post models.Post
		var publishDate sql.NullTime
		var tagLabels sql.NullString

		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Status, &publishDate, &tagLabels)
		if err != nil {
			fmt.Println("MASUKKKLAHHHHADINDAA==", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}

		if _, ok := postsMap[post.ID]; !ok {
			post.PublishDate = publishDate.Time
			postsMap[post.ID] = &post
		}

		if tagLabels.Valid {
			postsMap[post.ID].Tags = append(postsMap[post.ID].Tags, tagLabels.String)
		}
	}

	var posts []models.Post
	for _, post := range postsMap {
		posts = append(posts, *post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID < posts[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func SearchPostsByTag(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized: No Claims in Context", http.StatusUnauthorized)
		return
	}

	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "Bad Request: Missing tag parameter", http.StatusBadRequest)
		return
	}

	db := utils.ConnectDB()
	defer db.Close()

	var rows *sql.Rows
	var err error

	if claims.Role == "admin" || claims.Role == "user" {
		rows, err = db.Query(`SELECT p.id, p.title, p.content, p.status, p.publish_date, ARRAY_AGG(t.label) AS tags
		FROM posts p JOIN tags t ON p.id = t.posts_id
		WHERE t.label = $1
		GROUP BY p.id, p.title, p.content, p.status, p.publish_date`, tag)
	}

	if err != nil {
		log.Println("Error searching posts by tag:", err)
		http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var postsMap = make(map[int]*models.Post)

	for rows.Next() {
		var post models.Post
		var publishDate sql.NullTime
		var tagLabels sql.NullString

		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Status, &publishDate, &tagLabels)
		if err != nil {
			log.Println("Error scanning post:", err)
			http.Error(w, "Internal Server Error: Database Error", http.StatusInternalServerError)
			return
		}

		if _, ok := postsMap[post.ID]; !ok {
			if publishDate.Valid {
				post.PublishDate = publishDate.Time
			}
			postsMap[post.ID] = &post
		}

		if tagLabels.Valid {
			postsMap[post.ID].Tags = append(postsMap[post.ID].Tags, tagLabels.String)
		}
	}

	var posts []models.Post
	for _, post := range postsMap {
		posts = append(posts, *post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].ID < posts[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}
