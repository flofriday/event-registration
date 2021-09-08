package eventregistration

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func createRoutes(store *SqlStore, config *Config) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	/*
		// Static files and templates
		r.Get("/", handleHome(store, config))
		r.Get("/static/*", handleStatic())

		// REST API
		r.Route("/api", func(r chi.Router) {
			// Actions performed by users
			r.Post("/users", createUser(store))

			// Actions performend by admins
			if config.AdminEnable {
				r.Get("/users.csv", getUsersCSV(store))
				r.Get("/users.pdf", getUsersPDF(store))
				r.Get("/users.json", getUsersJSON(store))
				r.Get("/statistic", getStatistic(store))
			}
		})
	*/

	return r
}

func main() {
	// Load the config
	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Unable to load `config.toml`: %v", err.Error())
	}
	log.Printf("Successfully loaded config: %+v", config)

	// Setup the storage
	store, err := NewSqlStore(config.Filename)
	if err != nil {
		log.Fatalf("Unable to create a new storage `%s`: %v", config.Filename, err.Error())
	}

	// Setup the routes
	router := createRoutes(store, config)

	// Start the server
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
}
