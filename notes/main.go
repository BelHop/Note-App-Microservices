package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Note struct {
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	User        string    `bson:"user"`
	Date        time.Time `bson:"date"`
}

type NoteReturn struct {
	ID          primitive.ObjectID `bson:"_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	User        string             `bson:"user"`
	Date        time.Time          `bson:"date"`
}

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte("authentic"), nil)
}

/*
defaults sets default values for the Note struct.

It checks if the Date field is empty and assigns the current time if it is.
It also checks if the Description field is empty and assigns "N/A" if it is.
*/
func (nte *Note) defaults() {
	var t time.Time
	if nte.Date == t {
		nte.Date = time.Now()
	}
	if nte.Description == "" {
		nte.Description = "N/A"
	}
}

// main is the entry point of the Go program.
//
// main does the following:
// - Creates a new chi router.
// - Uses a logger middleware.
// - Registers a POST route "/new" with the "Create" handler.
// - Registers a route group "/find" with nested routes.
//   - Registers a GET route "/{user}/{title}" with the "Read" handler.
//   - Registers a GET route "/many/{user}/{title}" with the "ReadMany" handler.
//   - Registers a GET route "/{user}" with the "ReadAll" handler.
//
// - Registers a PUT route "/update/{user}/{title}" with the "Update" handler.
// - Registers a DELETE route "/delete/{user}/{title}" with the "Delete" handler.
// - Starts the HTTP server and listens on port 3000.
func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(jwtauth.Verifier(tokenAuth))
	r.Use(jwtauth.Authenticator(tokenAuth))

	r.Post("/new", Create)

	r.Route("/find", func(r chi.Router) {
		r.Get("/{user}/{title}", Read)
		r.Get("/many/{user}/{title}", ReadMany)
		r.Get("/{user}", ReadAll)
	})
	r.Put("/update/{user}/{title}", Update)
	r.Delete("/delete/{user}/{title}", Delete)
	http.ListenAndServe(":3000", r)
}

// Create creates a new note in the database.
//
// It takes in a http.ResponseWriter and a http.Request as parameters.
// There is no return value.
func Create(w http.ResponseWriter, r *http.Request) {
	var n Note
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	_, claims, _ := jwtauth.FromContext(r.Context())

	x, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(x, &n)
	n.defaults()
	if err != nil {
		log.Fatal(err)
	}
	if n.User == claims["username"] {
		clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)
		collection := client.Database(DB).Collection(Collection)
		_, err = collection.InsertOne(ctx, &n)
		if err != nil {
			log.Fatal("Could not insert doc")
		}
		w.Write([]byte(fmt.Sprintf("Successfully created '%s'", n.Title)))
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}

}

// Read handles the HTTP request and returns a note.
//
// The function takes in two parameters:
//   - w: an http.ResponseWriter object used to write the HTTP response
//   - r: an *http.Request object representing the HTTP request
//
// The function does not return any values.
func Read(w http.ResponseWriter, r *http.Request) {
	var n NoteReturn
	ctx := context.TODO()
	t := chi.URLParam(r, "title")
	u := chi.URLParam(r, "user")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	_, claims, _ := jwtauth.FromContext(r.Context())

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	if u == claims["username"] {
		filter := bson.D{{Key: "title", Value: t}, {Key: "user", Value: u}}
		collection := client.Database(DB).Collection(Collection)
		c, err := collection.Find(ctx, filter)
		if err != nil {
			log.Fatal("Could not find any matches")
		}
		defer c.Close(context.TODO())
		for c.Next(context.TODO()) {
			err := c.Decode(&n)
			if err != nil {
				log.Fatal("Could not decode a match")
			}
		}
		json.NewEncoder(w).Encode(&n)
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}
}

// ReadMany retrieves multiple documents from the database based on the provided title and user.
//
// It takes in the following parameters:
// - w: http.ResponseWriter - The response writer used to write the response back to the client.
// - r: *http.Request - The HTTP request received from the client.
//
// The function does not return any value.
func ReadMany(w http.ResponseWriter, r *http.Request) {
	var n []bson.M
	ctx := context.TODO()
	t := chi.URLParam(r, "title")
	u := chi.URLParam(r, "user")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	_, claims, _ := jwtauth.FromContext(r.Context())

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	if u == claims["username"] {
		filter := bson.D{{Key: "title", Value: bson.D{{"$regex", "^" + t}}}, {Key: "user", Value: u}}
		c, err := collection.Find(ctx, filter)
		if err != nil {
			log.Fatal("Could not find any matches")
		}
		defer c.Close(context.TODO())
		if err = c.All(ctx, &n); err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(&n)
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}
}

// ReadAll handles the HTTP request to read all documents for a specific user.
//
// It takes in the http.ResponseWriter and http.Request as parameters.
// It does not return anything.
func ReadAll(w http.ResponseWriter, r *http.Request) {
	var n []bson.M
	ctx := context.TODO()
	u := chi.URLParam(r, "user")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	_, claims, _ := jwtauth.FromContext(r.Context())

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	if u == claims["username"] {
		filter := bson.D{{Key: "user", Value: u}}
		c, err := collection.Find(ctx, filter)
		if err != nil {
			log.Fatal("Could not find any matches")
		}
		defer c.Close(context.TODO())
		if err = c.All(ctx, &n); err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(&n)
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}
}

// Update updates a note in the database based on the provided title and user.
//
// Parameters:
// - w: the http.ResponseWriter object for writing the response.
// - r: the *http.Request object representing the incoming request.
//
// Return type:
// This function does not return anything.
func Update(w http.ResponseWriter, r *http.Request) {
	var n Note
	ctx := context.TODO()
	t := chi.URLParam(r, "title")
	u := chi.URLParam(r, "user")
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	_, claims, _ := jwtauth.FromContext(r.Context())

	x, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(x, &n)
	n.defaults()
	if err != nil {
		log.Fatal(err)
	}

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	if u == claims["username"] {
		filter := bson.D{{Key: "title", Value: t}, {Key: "user", Value: u}}
		update := bson.D{{"$set", bson.D{{Key: "title", Value: n.Title}, {Key: "description", Value: n.Description}, {Key: "date", Value: n.Date}}}}
		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(fmt.Sprintf("Successfully updated '%s'", n.Title)))
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}
}

// Delete deletes a document from the specified collection in the MongoDB database.
//
// Parameters:
// - w: The http.ResponseWriter used to write the response back to the client.
// - r: The *http.Request containing the HTTP request details.
//
// Returns:
// The function does not return anything.
func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	t := chi.URLParam(r, "title")
	u := chi.URLParam(r, "user")
	filter := bson.D{{Key: "title", Value: t}, {Key: "user", Value: u}}

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)
	_, claims, _ := jwtauth.FromContext(r.Context())

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	if u == claims["username"] {
		_, err = collection.DeleteOne(ctx, filter)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(fmt.Sprintf("Successfully deleted '%s'", t)))
	} else {
		w.Write([]byte("You're not authorized to make this request"))
	}
}
