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

func (nte *Note) defaults() {
	var t time.Time
	if nte.Date == t {
		nte.Date = time.Now()
	}
	if nte.Description == "" {
		nte.Description = "N/A"
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
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

func Create(w http.ResponseWriter, r *http.Request) {
	var n Note
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}
	uri, DB, Collection := os.Getenv("URI"), os.Getenv("DB"), os.Getenv("Note")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

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
	_, err = collection.InsertOne(ctx, &n)
	if err != nil {
		log.Fatal("Could not insert doc")
	}
	w.Write([]byte(fmt.Sprintf("Successfully created '%s'", n.Title)))
}

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

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
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
}

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

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	filter := bson.D{{Key: "title", Value: bson.D{{"$regex", "^" + t}}}, {Key: "user", Value: u}}
	fmt.Print(filter)
	c, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal("Could not find any matches")
	}
	defer c.Close(context.TODO())
	if err = c.All(ctx, &n); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&n)
}

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

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	filter := bson.D{{Key: "user", Value: u}}
	fmt.Print(filter)
	c, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal("Could not find any matches")
	}
	defer c.Close(context.TODO())
	if err = c.All(ctx, &n); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&n)
}

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

	filter := bson.D{{Key: "title", Value: t}, {Key: "user", Value: u}}
	update := bson.D{{"$set", bson.D{{Key: "title", Value: n.Title}, {Key: "description", Value: n.Description}, {Key: "date", Value: n.Date}}}}
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte(fmt.Sprintf("Successfully updated '%s'", n.Title)))
}

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

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DB).Collection(Collection)

	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte(fmt.Sprintf("Successfully deleted '%s'", t)))
}
