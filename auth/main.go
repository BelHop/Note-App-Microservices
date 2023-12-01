package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	username    string `bson:"username"`
	password    string `bson:"username"`
	email       string `bson:"email"`
	DateOfBirth string `bson:"date of birth"`
}

type SignIn struct {
	username string `bson:"username"`
	password string `bson:"password"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/signup", SignUpHandler)
	r.Post("/signin", SignInHandler)
	r.Delete("/delete", DeleteUser)
	http.ListenAndServe(":3001", r)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}

	user, password, email, DateOfBirth := r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("email"), r.PostFormValue("DateOfBirth")

	u := User{
		username:    user,
		password:    password,
		email:       email,
		DateOfBirth: DateOfBirth,
	}

	uri := os.Getenv("URI")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("NotesApp").Collection("Users")

	_, err = collection.InsertOne(ctx, u)
	if err != nil {
		log.Fatal(err)
	}

	cookie := http.Cookie{
		Name:     "Authentication",
		Value:    u.username,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, &cookie)

	w.Write([]byte(fmt.Sprintf("User '%s' created!", u.username)))
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}

	var u User
	user, password := r.PostFormValue("username"), r.PostFormValue("password")
	filter := bson.D{{Key: "username", Value: user}, {Key: "password", Value: password}}

	uri := os.Getenv("URI")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("NotesApp").Collection("Users")

	c := collection.FindOne(ctx, filter)
	c.Decode(&u)

	cookie := http.Cookie{
		Name:     "Authentication",
		Value:    u.username,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, &cookie)
	w.Write([]byte("You're logged in!"))
}

// TODO: Make the following code work with the frontend of the service/implement an update function for users

// func UpdateUser(w http.ResponseWriter, r *http.Request) {
//   ctx := context.TODO()
// 	user, password, email, DateOfBirth := r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("email"), r.PostFormValue("DateOfBirth")
//   filter := bson.D{{Key: "username", Value: user}, {Key: "email", Value: email}}
// 	uri := "mongodb+srv://cluster-notes.72lwil1.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=./X509-cert-5068675552043678029.pem"
// 	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
//
// 	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)
//
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer client.Disconnect(ctx)
//
//   collection := client.Database("NotesApp").Collection("Users")
//   _, err = collection.UpdateOne(ctx, filter, )
// }

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}

	user := r.PostFormValue("username")
	filter := bson.D{{Key: "username", Value: user}}
	uri := os.Getenv("URI")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)

	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("NotesApp").Collection("Users")
	_, err = collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	w.Write([]byte(fmt.Sprintf("Successfully deleted account!")))
}
