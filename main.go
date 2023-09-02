package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rajagennu/rssfeeder/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ReadynessResponse struct {
	Status string `json:"status"`
}

func main() {
	PORT := getPort()
	dbUrl := getDBUrl()
	log.Println("Read port from .env file ", PORT)
	log.Println("Read DB conn string from .env file ", dbUrl)

	// open connection to the database
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error while trying to connect with database ", err.Error())
	}

	apiCfg := apiConfig{}
	dbQueries := database.New(db)
	apiCfg.DB = dbQueries

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	router.Mount("/v1", v1Router())
	server := &http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("error while starting the server %s %q ", PORT, err)
	}

}

func v1Router() chi.Router {
	router := chi.NewRouter()
	router.Get("/readiness", handleReadyNess)
	router.Get("/err", handleErrorResponse)
	router.Post("/users", createUser)

	return router
}

func getPort() string {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error while try to load .env file, Please check")
		return ""
	}
	key := "PORT"
	return os.Getenv(key)
}

func getDBUrl() string {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error while trying to load .env file, PLease check")
		return ""
	}
	key := "DB"
	return os.Getenv(key)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.WriteHeader(status)
	w.Header().Set("content-type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithJSON(w, 500, "Internal Server Error")
		return
	}
	w.Write(data)

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	err := ErrorResponse{
		Error: msg,
	}
	data, _ := json.Marshal(err)
	w.Write(data)

}

func handleReadyNess(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, 200, ReadynessResponse{Status: "ok"})
}

func handleErrorResponse(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Id         string    `json:"id"`
		Name       string    `json:"name"`
		Created_at time.Time `json:"created_at"`
		Updated_at time.Time `json:"updated_at"`
	}

	var newUser user
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	err := d.Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if newUser.Name == "" {
		http.Error(w, "Name is empty ", http.StatusBadRequest)
		return
	}

	newUser.Id = uuid.NewString()
	newUser.Updated_at = time.Now()
	newUser.Created_at = time.Now()

	respondWithJSON(w, 201, newUser)

}
