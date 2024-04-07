package main

import (
    "context"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/go-redis/redis/v8"
    "crypto/sha1"
    "encoding/hex"
    "io"
    "time"
)

var redisClient *redis.Client
var ctx = context.Background()

func main() {
    // Initialize Redis client
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379", // Redis server address
        Password: "",               // No password set
        DB:       0,                // Use default DB
    })

    // Initialize HTTP server and router
    r := mux.NewRouter()
    r.HandleFunc("/{shortUrl}", RedirectHandler).Methods("GET")
    r.HandleFunc("/shorten", ShortenHandler).Methods("POST")

    http.Handle("/", r)
    fmt.Println("Server is listening on port 8080...")
    http.ListenAndServe(":8080", nil)
}

// RedirectHandler redirects to the original URL
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    shortUrl := vars["shortUrl"]

    // Retrieve the original URL from Redis
    originalUrl, err := redisClient.Get(ctx, shortUrl).Result()
    if err == redis.Nil {
        http.NotFound(w, r)
        return
    } else if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, originalUrl, http.StatusFound)
}

// ShortenHandler shortens the URL
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    originalUrl := r.Form.Get("url")

    // Generate a short URL
    shortUrl := generateShortLink(originalUrl)

    // Store the original URL in Redis with the short URL as the key
    err := redisClient.Set(ctx, shortUrl, originalUrl, 24*time.Hour).Err()
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Shortened URL: %s/%s", r.Host, shortUrl)
}

// generateShortLink generates a short link hash
func generateShortLink(originalUrl string) string {
    hasher := sha1.New()
    io.WriteString(hasher, originalUrl)
    sha := hasher.Sum(nil)
    shortUrl := hex.EncodeToString(sha)[:8] // Use the first 8 characters
    return shortUrl
}
