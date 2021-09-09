package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func handleHome(store *SqlStore, config *Config) http.HandlerFunc {
	// Setup some templates and files, this will only run once and can be used
	// concurrently by the function that will be returned.
	successFilename := path.Join("templates", config.Template, "success.html")
	successTemplate := template.Must(template.ParseFiles(successFilename))

	registerFilename := path.Join("templates", config.Template, "register.html")
	registerDocument, err := os.ReadFile(registerFilename)
	if err != nil {
		panic("Unable to find template register.html")
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		// Extract the cookie like a blue monster and see if the user is logged
		// in
		cookie, err := r.Cookie("auth")
		if err != nil {
			// There is no available cookie so just return the registration
			//page
			log.Printf("User has no cookie: %s", err.Error())
			rw.WriteHeader(http.StatusOK)
			rw.Write(registerDocument)
			return
		}

		// So the user has the right cookie, so lets load their contact data
		// and show them that.
		user, err := store.getByUuid(cookie.Value)
		if err != nil {
			// The user could not be loaded so maybe his cookie is bad or the
			// database got reset since they logged in.
			// Anyway, we will just gonna send them the register page and
			// reset their cookie.
			log.Printf("No user with UUID %s could loaded: %s", cookie.Value, err.Error())
			cookie := http.Cookie{
				Name:     "auth",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
			}
			http.SetCookie(rw, &cookie)
			rw.WriteHeader(http.StatusOK)
			rw.Write(registerDocument)
			return
		}

		// Ok the user is really logged in so show them who they are:
		rw.WriteHeader(http.StatusOK)
		successTemplate.Execute(rw, user)
	}
}

func createUser(store *SqlStore) http.HandlerFunc {
	// Create a new struct to match the input of the request
	type UserInput struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		// Parse the input
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("The body cannot be read: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var input UserInput
		err = json.Unmarshal(body, &input)
		if err != nil {
			log.Printf("The request body is not valid JSON: %s", err.Error())
			http.Error(rw, "Unprocessable Entity", http.StatusUnprocessableEntity)
			return
		}

		// TODO: do validation here

		// Create a new user object
		uuid, err := createUuid()
		if err != nil {
			log.Printf("Unable to create a new UUID: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		user := User{
			UUID:      uuid,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			Email:     input.Email,
			Phone:     input.Phone,
			CreatedAt: time.Now().UTC(),
		}

		// Add the user to the storage
		err = store.add(user)
		if err != nil {
			log.Printf("Unable to insert the user: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Successfully registered the user, set them a cookie to indicate that
		// they are logged in.
		cookie := http.Cookie{
			Name:     "auth",
			Value:    user.UUID,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour * 24 * 3),
			HttpOnly: true,
		}
		http.SetCookie(rw, &cookie)
		rw.WriteHeader(http.StatusOK)
	}
}

func getUsersCSV(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func getUsersPDF(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func getUsersJSON(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func getStatistic(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func createRoutes(store *SqlStore, config *Config) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Static files and templates
	r.Get("/", handleHome(store, config))
	filesDir := http.Dir("static")
	FileServer(r, "/static", filesDir)

	// REST API
	r.Route("/api", func(r chi.Router) {
		// Actions performed by users
		r.Post("/users", createUser(store))

		// Actions performend by admins
		// TODO: add admin middleware
		if config.AdminEnable {
			r.Get("/users.csv", getUsersCSV(store))
			r.Get("/users.pdf", getUsersPDF(store))
			r.Get("/users.json", getUsersJSON(store))
			r.Get("/statistic", getStatistic(store))
		}
	})

	return r
}

func main() {
	// Load the config
	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Unable to load `config.toml`: %v", err.Error())
	}
	log.Printf("Successfully loaded config: %#v", *config)

	// Setup the storage
	store, err := NewSqlStore(config.Filename)
	if err != nil {
		log.Fatalf("Unable to create a new storage `%s`: %v", config.Filename, err.Error())
	}
	log.Printf("Successfully initialized a store in: %s", config.Filename)

	// Setup the routes
	router := createRoutes(store, config)

	// Start the server
	log.Printf("Started server at port %d", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
}
