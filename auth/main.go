package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Username    string `bson:"username"`
	Password    string `bson:"password"`
	Email       string `bson:"email"`
	DateOfBirth string `bson:"date of birth"`
}

type SignIn struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/auth", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"https://localhost:5173", "http://localhost:5173"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"POST"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
		r.Post("/signup", SignUpHandler)
		r.Post("/signin", SignInHandler)
	})
	r.Delete("/delete", DeleteUser)
	http.ListenAndServe(":3001", r)
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}

	// user, password, email, DateOfBirth := r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("email"), r.PostFormValue("DateOfBirth")
	//
	// u := User{
	// 	Username:    user,
	// 	Password:    password,
	// 	Email:       email,
	// 	DateOfBirth: DateOfBirth,
	// }
	var u User
	json.NewDecoder(r.Body).Decode(&u)

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

	w.Header().Set("Authorization", JWTcreate(u.Username, u.Password))

}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Could not load env variable")
	}

	var u SignIn
	json.NewDecoder(r.Body).Decode(&u)
	filter := bson.D{{Key: "username", Value: u.Username}, {Key: "password", Value: u.Password}}
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
	if c.Err() != nil {
		w.Write([]byte("No matching documents\n u"))
	}
	c.Decode(&u)
	w.Header().Set("Authorization", JWTcreate(u.Username, u.Password))
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

func JWTcreate(username, password string) string {
	token := jwtauth.New("HS256", []byte("authentic"), nil)
	_, string, _ := token.Encode(map[string]interface{}{"username": username, "password": password})
	return string
}
