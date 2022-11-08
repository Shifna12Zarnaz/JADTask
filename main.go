package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	//     "os"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	// "github.com/kelseyhightower/envconfig"
)

type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

var db *sql.DB
var err error

//	type environ struct{
//		dbUsername     string `envconfig:"DB_USER" default:"root"`
//	  dbPassword     string `envconfig:"DB_PASSWORD" default:"shifna"`
//	  dbname         string  `envconfig:"DB_NAME" default:"newdb"`
//		dbHost         string  `envconfig:"DB_HOST" default:"tcp(127.0.0.1:3306)"`
//	}
//
// const (
//
//		`envconfig:"NUMBER_ONE" default:"1"`
//	 DB_USER     = `envconfig:"DB_USER" default:"root"`
//
// DB_PASSWORD = `envconfig:"DB_PASSWORD" default:"shifna"`
//
//	DB_NAME     = "newdb"
//
// )
func main() {
	//	env := &environ{}
	//	db, err = sql.Open("mysql", "root:shifna@tcp(127.0.0.1:3306)/newdb")
	//db, err = sql.Open("mysql", "root:shifna/newdb")
	//	dbHost := os.Getenv("DB_HOST")
	//	dbUsername := os.Getenv("DB_USERNAME")
	//	dbPassword := os.Getenv("DB_PASSWORD")
	//	dbname := os.Getenv("DB_NAME")
	os.Setenv("DB_HOST", "tcp(127.0.0.1:3306)")
	os.Setenv("DB_USERNAME", "root")
	os.Setenv("DB_PASSWORD", "shifna")
	os.Setenv("DB_NAME", "newdb")
	dbHost := os.Getenv("DB_HOST")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dsn := dbUsername + ":" + dbPassword + "@" + dbHost + "/" + dbname + "?charset=utf8"
	//    dsn := env.dbUsername + ":" + env.dbPassword + "@" + env.dbHost + "/" + env.dbname + "?charset=utf8"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	//	db, err = sql.Open("mysql", "root:shifna@tcp(127.0.0.1:3306)/newdb")
	//	if err != nil {
	//		panic(err.Error())
	//	}
	defer db.Close()
	router := mux.NewRouter()
	router.HandleFunc("/user", getPosts).Methods("GET")
	router.HandleFunc("/useradd", createPost).Methods("POST")
	router.HandleFunc("/user/{id}", getPost).Methods("GET")
	router.HandleFunc("/health", HealthCheckHandler).Methods("GET")
	http.ListenAndServe(":8000", router)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var posts []Post
	result, err := db.Query("SELECT id, title from posts")
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	for result.Next() {
		var post Post
		err := result.Scan(&post.ID, &post.Title)
		if err != nil {
			panic(err.Error())
		}
		posts = append(posts, post)
	}
	json.NewEncoder(w).Encode(posts)
}
func createPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("INSERT INTO posts(title) VALUES(?)")
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	title := keyVal["title"]
	_, err = stmt.Exec(title)
	if err != nil {
		panic(err.Error())
	}
	fmt.Fprintf(w, "New post was created")
}
func getPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := db.Query("SELECT id, title FROM posts WHERE id = ?", params["id"])
	if err != nil {
		panic(err.Error())
	}
	defer result.Close()
	var post Post
	for result.Next() {
		err := result.Scan(&post.ID, &post.Title)
		if err != nil {
			panic(err.Error())
		}
	}
	json.NewEncoder(w).Encode(post)
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusOK)

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	//	io.WriteString(w, `{"alive": true}`)
	resp, err := http.Get("http://localhost:8000/user")
	if err != nil {
		log.Fatalf("HTTP GET request failed, %v\n", err)
	}
	fmt.Fprintf(w, "<h1>Health check is done  %v</h1>", resp.Status)
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
	scanner := bufio.NewScanner(resp.Body)
	for i := 0; scanner.Scan() && i < 5; i++ {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Body read failed: %v\n", err)
	}
}
