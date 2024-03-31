package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

var days []string = []string{"Maanantai", "Tiistai", "Keskiviikko", "Torstai", "Perjantai", "Lauantai", "Sunnuntai"}
var mealNames []string = []string{"Lounas", "Illallinen"}

type Meal struct {
	Id        int
	Locked    bool
	Recipe    *Recipe
	MealLabel string
	Clear     bool
	FlipUrl   string
	ClearUrl  string
}

type Generation struct {
	CurKey string
	Meals  []Meal
}

func (g Generation) encode() string {
	initial := ""
	for _, meal := range g.Meals {
		id := meal.Id
		if meal.Clear {
			id = 424242
		}
		if meal.Locked {
			id *= -1
		}
		initial += fmt.Sprintf("%d;", id)
	}
	return initial
}

// Do db queries to fetch actual recipies with the ids
func (g *Generation) fillRecipes() error {
	for i, meal := range g.Meals {
		if meal.Clear {
			continue
		}
		r, e := GetRecipeById(meal.Id)
		if e != nil {
			return e
		}
		g.Meals[i].Recipe = &r
	}
	return nil
}

func getRandomRecipeId(rl []Recipe) int {
	num := rand.Intn(len(rl))
	return rl[num].Id
}

func (g *Generation) generate(n int) {
	oldGen := g.Meals
	newGen := make([]Meal, n)

	// Get all recipes.
	arr := GetRecipes()

	for i := 0; i < n; i++ {
		if i < len(oldGen) {
			if oldGen[i].Locked {
				newGen[i] = oldGen[i]
			} else {
				newGen[i] = Meal{Id: getRandomRecipeId(arr), Locked: false}
			}
		} else {
			newGen[i] = Meal{Id: getRandomRecipeId(arr), Locked: false}
		}
	}

	g.Meals = newGen
}

func decode(s string) (g Generation) {
	fields := strings.Split(s, ";")
	log.Println("fields: ", fields)
	gen := Generation{}
	for _, field := range fields {
		if field == "" {
			continue
		}
		id, err := strconv.Atoi(field)
		if err != nil {
			log.Println("was unable to parse field err: ", err)
			continue
		}
		meal := Meal{}
		if id == 424242 || id == -424242 {
			meal.Clear = true
			meal.Recipe = &Recipe{}
			meal.Recipe.Name = "Empty"
		}
		if id < 0 {
			meal.Locked = true
			id *= -1
		}
		meal.Id = id
		gen.Meals = append(gen.Meals, meal)
	}
	log.Println(gen)
	return gen
}

func (g *Generation) genFlipUrls() {
	for i := 0; i < len(g.Meals); i++ {
		g.Meals[i].Locked = !g.Meals[i].Locked // Flip
		fmt.Println(g.Meals[i].Locked)
		g.Meals[i].FlipUrl = "/gen/" + g.encode()
		oc := g.Meals[i].Clear
		g.Meals[i].Clear = true
		g.Meals[i].ClearUrl = "/gen/" + g.encode()
		g.Meals[i].Clear = oc
		g.Meals[i].Locked = !g.Meals[i].Locked // Flip back
	}
}

func (g *Generation) genMealLabels(start int) {
	for i := 0; i < len(g.Meals); i++ {
		g.Meals[i].MealLabel = days[((i+start)/2)%7] + " - " + mealNames[(i+start)%2]
	}
}

func addGenerateRoute(r *chi.Mux) {
	r.Get("/gen/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")

		startMeal := 0
		startingMealQString := r.URL.Query().Get("startMeal") // Query parameter from 0 to 13 representing the meal the list should start on.
		startingMealQ, err := strconv.Atoi(startingMealQString)
		if err == nil && startingMealQ >= 0 && startingMealQ <= 13 {
			startMeal = startingMealQ
		}

		if key == "none" { // If key is empty, gen a 14 recipe key and redir to that
			gen := Generation{}
			gen.generate(14)
			s := gen.encode()
			http.Redirect(w, r, "/gen/"+s, http.StatusSeeOther)
			return
		}
		gen := decode(key)
		if gen.fillRecipes() != nil {
			http.Error(w, "was not able to get the recipes specified in the generation", http.StatusInternalServerError)
		}

		gen.genFlipUrls()
		gen.genMealLabels(startMeal)
		gen.CurKey = key

		log.Println("Has key!")

		views.ExecuteTemplate(w, "generation.tmpl", gen)
	})

	r.Get("/gen/re/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		gen := decode(key)
		gen.generate(14)
		s := gen.encode()
		http.Redirect(w, r, "/gen/"+s, http.StatusSeeOther)
	})
}
