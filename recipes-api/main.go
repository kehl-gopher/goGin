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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipes []Recipe
var ctx context.Context
var err error
var client *mongo.Client
var collection *mongo.Collection

func init() {
	recipes = make([]Recipe, 0)

	ctx = context.Background()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		log.Fatal(err)
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB")
	// LoadData(client)
}

func LoadData(client *mongo.Client) {
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

	var listOfRecipes []interface{}

	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	var ctx = context.Background()
	insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
	// defer f.Close()
}

type Recipe struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`

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

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"tags", recipe.Tags},
		{"ingredients", recipe.Ingredients},
		{"instructions", recipe.Instructions},
	}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe updated"})
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

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectID, _ := primitive.ObjectIDFromHex(id)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
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
	cur, err := collection.Find(ctx, bson.M{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	recipes := make([]Recipe, 0)

	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	// router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)
	router.Run()
}
