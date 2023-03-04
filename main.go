package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func routes() http.Handler {
	mux := chi.NewMux()
	mux.HandleFunc("/put", write)

	return mux
}

type user struct {
	User string `json:"user"`
	Age  int    `json:"age"`
}

func write(w http.ResponseWriter, r *http.Request) {

	log.Println("write()")
	// read the JSON from the string
	var u user
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&u); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if err := writeUser(u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Written succesfully"))
}

func writeUser(u user) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:password@mongodb:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	log.Println("Connected")
	collection := client.Database("my-db").Collection("users")
	if collection != nil {
		log.Println("got users")
	}
	_, err = collection.InsertOne(ctx, bson.M{"name": u.User, "age": u.Age})
	if err != nil {
		log.Fatalf("%+v insert failure %+v", u, err)
	}
	return nil
}

func client(usr string, age int) {
	var c http.Client
	u := user{
		User: usr,
		Age:  age,
	}

	bs, err := json.Marshal(u)
	if err != nil {
		log.Fatal(err)
	}

	b := bytes.NewBuffer(bs)
	log.Println(string(bs))
	r, err := http.NewRequest(http.MethodPost, "http://localhost:80/put", b)
	if err != nil {
		log.Fatal("unable to create request " + err.Error())
	}
	r.Header.Set("content-type", "application/json")

	log.Println("HTTP Post...")
	resp, err := c.Do(r)
	if err != nil {
		log.Fatal("HTTP send request failed " + err.Error())
	}
	defer resp.Body.Close()
	log.Println("Status returned: ", resp.StatusCode)

}

func main() {
	// Arguments expected:
	//	server:port
	//  client:user:age

	if len(os.Args) <= 0 {
		log.Fatal("Too few args..")
	}

	var port string
	var usr string
	var age int

	s := strings.Split(os.Args[1], ":")
	switch strings.ToUpper(s[0]) {
	case "SERVER":
		port = s[1]
	case "CLIENT":
		usr = s[1]
		age, _ = strconv.Atoi(s[2])
	default:
		log.Fatal(s[0] + " invalid param value")
	}
	if port != "" {
		mux := routes()
		log.Println("Listening on port ", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatal(err)
		}
		return
	}
	client(usr, age)
}
