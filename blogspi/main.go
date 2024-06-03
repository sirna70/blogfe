package main

import (
	"fmt"
	"log"
	"net/http"

	h "blogspi/cmd"
	"blogspi/handlers"
	"blogspi/handlers/middleware"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.Use(h.CorsMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "WELCOME TO BLOGGER SIMPLE")
	})
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/register", handlers.Register).Methods("POST")

	postRouter := r.PathPrefix("/posts").Subrouter()
	postRouter.Use(middleware.AuthMiddleware)
	postRouter.HandleFunc("", handlers.CreatePost).Methods("POST")
	postRouter.HandleFunc("/update", handlers.UpdatePost).Methods("PUT")
	postRouter.HandleFunc("/publish", handlers.PublishPost).Methods("PUT")
	postRouter.HandleFunc("/delete", handlers.DeletePost).Methods("DELETE")
	postRouter.HandleFunc("/get", handlers.GetPosts).Methods("GET")
	postRouter.HandleFunc("/search", handlers.SearchPostsByTag).Methods("GET")

	log.Fatal(http.ListenAndServe(":9090", r))
}
