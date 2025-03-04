// Recipes API
// This is a sample recipes API. you can find out more about this API at https://github.com/PacktPublishing/Building-Distributed-Applications-in-Gin
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Darasimi <kelanidarasimi9@gmail.com>
// Consumes:
// - application/json
// Produces:
// - application/json
// swagger:meta
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	LoadData()
}

func LoadData() {
	// f, err := os.Open("recipes.json")
	data, err := os.ReadFile("recipes.json")

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	err = json.Unmarshal(data, &recipes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	// defer f.Close()
}

type Recipe struct {
	ID string `json:"id"`

	// The name of the recipe
	// example: "Spaghetti Carbonara"
	Name string `json:"name"`
	// List of tags for categorization
	// example: ["Italian", "Pasta", "Dinner"]
	Tags []string `json:"tags"`
	// List of ingredients
	// example: ["Eggs", "Parmesan Cheese", "Bacon", "Spaghetti"]
	Ingredients []string `json:"ingredients"`
	// Step-by-step cooking instructions
	// example: ["Boil pasta", "Cook bacon", "Mix eggs with cheese", "Combine everything"]
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for i := 0; i < len(recipes); i++ {
		found := false
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, recipes[i])
		}
	}
	c.JSON(http.StatusOK, listOfRecipes)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Updates a recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe to update
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//   '200':
//     description: successful operation
//   '400':
//     description: invalid data
//   '404':
//     description: invalid recipe ID

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	index := -1

	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	recipe.ID = id
	recipes[index] = recipe
	fmt.Println(index)
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation POST /recipes recipes newRecipe
// Creates a new recipe
// ---
// consumes:
//   - application/json
//
// produces:
//   - application/json
//
// parameters:
//   - in: body
//     name: recipe
//     description: The recipe to create
//     required: true
//     schema:
//     $ref: "#/definitions/Recipe"
//
// responses:
//
//	'201':
//	  description: successful operation
//	'400':
//	  description: invalid request data
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe

	if err := c.ShouldBindBodyWithJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	recipes = append(recipes, recipe)

	c.JSON(http.StatusCreated, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	index := -1

	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}

	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	recipes = slices.Delete(recipes, index, index+1)
	c.JSON(http.StatusOK, gin.H{"message": "recipe deleted"})
}

// swagger:operation GET /recipes recipes listRecipes
// Returns a list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	 description: successful operation
//	schema:
//	type: array
//
// items:
// $ref: '#/definitions/Recipe'
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)
	router.Run()
}
