package main

import (
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
)

// Operator is a ...
type Operator struct {
	connection redis.Conn
}

func (o *Operator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Init Operator connection if needed
	if o.connection == nil {
		o.connect()
	}

	if r.Method == "POST" {
		o.create(w, r)
	} else {
		o.lookup(w, r)
	}
}

// Setup redis connection
// Defaults to connecting on local redis (port 6379)
// This can be customised using REDIS_PORT
func (o *Operator) connect() {
	port := envOrDefault("REDIS_PORT", "6379")
	conn, err := redis.Dial("tcp", ":"+port)
	if err != nil {
		panic(err) // Can't do much without a redis connection
	}

	o.connection = conn
}

// Create token in redis, with url as value
// Returns 200 if token created successfully
// Returns 400 bad request if not
func (o *Operator) create(w http.ResponseWriter, r *http.Request) {
	token := o.parseToken(r)
	url := r.FormValue("url")

	reply, _ := redis.Int(o.connection.Do("SETNX", token, url))
	if reply != 1 {
		http.Error(w, "Token already used", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Lookup url by token
// Redirects to url if found
// Returns 404 not found if not
func (o *Operator) lookup(w http.ResponseWriter, r *http.Request) {
	token := o.parseToken(r)

	url, _ := redis.String(o.connection.Do("GET", token))
	if url == "" {
		http.Error(w, "Token not found", http.StatusNotFound)
		return
	}

	log.Printf("Connecting %v to %v", token, url)
	http.Redirect(w, r, url, 301)
}

func (o *Operator) parseToken(r *http.Request) string {
	return r.URL.Path[1:] // Strip leading slash
}