package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const cookieName = "event-registration-auth"

func deleteCookie(rw http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(rw, &cookie)
}

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
		cookie, err := r.Cookie(cookieName)
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
		user, err := store.GetByUuid(cookie.Value)
		if err != nil {
			// The user could not be loaded so maybe his cookie is bad or the
			// database got reset since they logged in.
			// Anyway, we will just gonna send them the register page and
			// reset their cookie.
			log.Printf("No user with UUID %s could loaded: %s", cookie.Value, err.Error())
			deleteCookie(rw)
			rw.WriteHeader(http.StatusOK)
			rw.Write(registerDocument)
			return
		}

		// Ok the user is really logged in so show them who they are:
		rw.WriteHeader(http.StatusOK)
		successTemplate.Execute(rw, user)
	}
}

func handleAdmin(store *SqlStore, config *Config) http.HandlerFunc {
	// The datastructures that are used to render the admin template
	type Statistic struct {
		UsersTotal int
		LastUsers  []User
	}
	type AdminData struct {
		Config    *Config
		Statistic *Statistic
	}

	// Setup some templates and files, this will only run once and can be used
	// concurrently by the function that will be returned.
	adminFilename := path.Join("templates", config.Template, "admin.html")
	adminTemplate := template.Must(template.ParseFiles(adminFilename))

	return func(rw http.ResponseWriter, r *http.Request) {

		// This could be more efficient in SQL solved
		users, err := store.GetLastN(10)
		if err != nil {
			log.Printf("Unable to load last n users: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		total, err := store.Count()
		if err != nil {
			log.Printf("Unable to count users: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := AdminData{
			Config: config,
			Statistic: &Statistic{
				UsersTotal: total,
				LastUsers:  users,
			},
		}

		adminTemplate.Execute(rw, data)
	}
}

// reset deletes the auth cookie of any registered user so that they can
// register again. This a hack so that one phone can sign up multiple attendees
func reset() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		deleteCookie(rw)
		http.Redirect(rw, r, "/", http.StatusMovedPermanently)
	}
}

func userValidater() func(*User) error {
	// This regex is not perfect and is probably too lax. However, for our
	// task at hand it is good enough. If you want to read more about email
	// regexes, I can only recommend: https://www.regular-expressions.info/email.html
	emailRegexp := regexp.MustCompile(`^.+@.+\.[A-Za-z]{2,}$`)

	// Yes this phone regex is also very liberal. I think it is save to assume
	// that there are 6 characters at least.
	phoneRegexp := regexp.MustCompile(`^\+?[0-9\-\s\(\)]{6,}$`)

	notEmptyRegexp := regexp.MustCompile("^.+$")

	return func(user *User) error {
		msg := make([]string, 0)

		if !notEmptyRegexp.MatchString(user.FirstName) {
			msg = append(msg, "first name is empty")
		}

		if !notEmptyRegexp.MatchString(user.LastName) {
			msg = append(msg, "second name is empty")
		}

		if !phoneRegexp.MatchString(user.Phone) {
			log.Println(user.Phone)
			msg = append(msg, "phonenumber is invalid")
		}

		if !emailRegexp.MatchString(user.Email) {
			msg = append(msg, "email is invalid")
		}

		if len(msg) == 0 {
			return nil
		}

		return errors.New(strings.Join(msg, ", "))
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

	validateUser := userValidater()

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

		// Create a new user object
		uuid, err := createUuid()
		if err != nil {
			log.Printf("Unable to create a new UUID: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		user := User{
			UUID:      uuid,
			FirstName: strings.TrimSpace(input.FirstName),
			LastName:  strings.TrimSpace(input.LastName),
			Email:     strings.TrimSpace(input.Email),
			Phone:     strings.TrimSpace(input.Phone),
			CreatedAt: time.Now(),
		}

		// Validate that the input is correct
		if err = validateUser(&user); err != nil {
			log.Printf("Validation failed: %s", err.Error())
			msg := "Your input failed the validation: " + err.Error()
			http.Error(rw, msg, http.StatusBadRequest)
			return
		}

		// Add the user to the storage
		err = store.Add(user)
		if err != nil {
			log.Printf("Unable to insert the user: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Successfully registered the user, set them a cookie to indicate that
		// they are logged in.
		cookie := http.Cookie{
			Name:     cookieName,
			Value:    user.UUID,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour * 24 * 3),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(rw, &cookie)
		rw.WriteHeader(http.StatusCreated)
	}
}

func deleteUser(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		uuid := chi.URLParam(r, "uuid")

		err := store.DeleteByUuid(uuid)
		if err != nil {
			log.Printf("Unable to delete user '%s': %s", uuid, err.Error())
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func getUsersCSV(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Load all users
		users, err := store.GetAll()
		if err != nil {
			log.Printf("Unable to load all users: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create a new csv writer
		rw.Header().Add("Content-Type", "text/csv")
		rw.WriteHeader(200)
		rw.Header().Add("Content-Disposition", "attachment")
		csvWriter := csv.NewWriter(rw)

		// Write the first line
		csvWriter.Write([]string{
			"First Name", "Last Name", "Email", "Phonenumber", "Registered at",
		})

		// Write all users
		for _, user := range users {
			csvWriter.Write([]string{
				user.FirstName,
				user.LastName,
				user.Email,
				user.Phone,
				user.CreatedAt.Format("2006-01-02 15:04:05 MST"),
			})

		}
		csvWriter.Flush()
	}
}

func getUsersJSON(store *SqlStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Load all users
		users, err := store.GetAll()
		if err != nil {
			log.Printf("Unable to load all users: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert to JSON
		data, err := json.Marshal(users)
		if err != nil {
			log.Printf("Unable encode users to json: %s", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}
}

func createRoutes(store *SqlStore, config *Config) chi.Router {
	// Setup the router and the admin middleware (requires a basicAuth password)
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	message := "Login as a admin"
	credentials := map[string]string{"admin": config.AdminPassword}
	adminMiddleware := middleware.BasicAuth(message, credentials)

	// Templates
	r.Get("/", handleHome(store, config))
	r.Get("/reset", reset())

	// Admin webinterface
	if config.AdminEnable {
		r.Route("/admin", func(r chi.Router) {
			r.Use(adminMiddleware)
			r.Get("/", handleAdmin(store, config))
		})
	}

	// REST API
	r.Route("/api", func(r chi.Router) {
		// Actions performed by users
		r.Post("/users", createUser(store))

		// Actions performend by admins
		if config.AdminEnable {
			r.Route("/", func(r chi.Router) {
				r.Use(adminMiddleware)

				r.Delete("/users/{uuid}", deleteUser(store))
				r.Get("/users.csv", getUsersCSV(store))
				r.Get("/users.json", getUsersJSON(store))
			})
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
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router)
	if err != nil {
		log.Fatalf("Starting the server failed: %s", err.Error())
	}
}
