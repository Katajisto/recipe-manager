package main

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var views = template.Must(template.ParseGlob("templates/*.tmpl"))

type RecipeListInput struct {
	Recipes []Recipe
	Oob     bool
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 10))

	authRouter := chi.NewRouter()
	authRouter.Use(ConfigMiddleware)
	authRouter.Use(AuthMiddleware)

	createDb()
	defer closeDb()

	workDir, _ := os.Getwd()
	fileDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", fileDir)

	authRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := views.ExecuteTemplate(w, "index.tmpl", RecipeListInput{Recipes: GetRecipes(), Oob: false})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Post("/doinit", func(w http.ResponseWriter, r *http.Request) {
		passwd := r.FormValue("password")
		hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = CreateConfig(string(hash))

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.Post("/dologin", func(w http.ResponseWriter, r *http.Request) {
		passwd := r.FormValue("password")
		config := GetConfig()
		if bcrypt.CompareHashAndPassword([]byte(config.PasswordHash), []byte(passwd)) == nil {
			t := make([]byte, 50)
			rand.Read(t)

			tokenString := hex.EncodeToString(t)

			err := CreateToken(tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			cookie := http.Cookie{
				Name:     "auth",
				Value:    tokenString,
				Path:     "/",
				MaxAge:   0,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			}

			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			views.ExecuteTemplate(w, "login.tmpl", "Login failed!")
			return
		}
	})

	authRouter.Get("/recipeAdd", func(w http.ResponseWriter, r *http.Request) {
		err := views.ExecuteTemplate(w, "addRecipe.tmpl", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	authRouter.Get("/recipes", func(w http.ResponseWriter, r *http.Request) {
		err := views.ExecuteTemplate(w, "recipesList.tmpl", RecipeListInput{Recipes: GetRecipes(), Oob: false})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	authRouter.Get("/recipes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		idInt, err := strconv.Atoi(id)

		recipe, err := GetRecipeById(idInt)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
		}

		err = views.ExecuteTemplate(w, "recipe.tmpl", recipe)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	authRouter.Delete("/recipes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		idInt, err := strconv.Atoi(id)

		err = DeleteRecipeById(idInt)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusNotFound)
		}

		err = views.ExecuteTemplate(w, "centralPanel.tmpl", nil)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = views.ExecuteTemplate(w, "recipesList.tmpl", RecipeListInput{Recipes: GetRecipes(), Oob: true}) // Regen the recipeList for the OOB swap
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	authRouter.Post("/recipe", func(w http.ResponseWriter, r *http.Request) {
		nr := Recipe{Name: r.FormValue("name"), Description: r.FormValue("desc")}
		err := nr.Add()
		if err != nil {
			log.Println("error: ", err)
		}
		err = views.ExecuteTemplate(w, "recipesList.tmpl", RecipeListInput{Recipes: GetRecipes(), Oob: false})
		if err != nil {
			log.Println("error: ", err)
		}
	})

	addGenerateRoute(authRouter)

	r.Mount("/", authRouter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Starting server at port: ", port)

	http.ListenAndServe(":"+port, r)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
