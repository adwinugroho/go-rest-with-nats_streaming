package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	//stan "github.com/nats-io/stan.go"
)

const (
	port = ":8000"
)

//models customer
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

//end models customer

//models notif
type Notif struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

//end models notif

//init var as  a slice struct
var users []User
var notifs []Notif

//handlers customer
func getListUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) //get param

	//Loop through books and find id
	for _, item := range users {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&User{})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)
	user.ID = strconv.Itoa(rand.Intn(10000000))
	users = append(users, user)
	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range users {
		if item.ID == params["id"] {
			users = append(users[:index], users[index+1:]...)
			var user User
			_ = json.NewDecoder(r.Body).Decode(&user)
			user.ID = params["id"]
			users = append(users, user)
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	json.NewEncoder(w).Encode(users)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range users {
		if item.ID == params["id"] {
			users = append(users[:index], users[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(users)
}

//end handlers customer

func main() {
	//initial data -implement database
	users = append(users, User{ID: "1", Username: "Pangeran", Email: "bangpange@gmail.com"})
	//end initial data

	//init router
	mRoute := mux.NewRouter()

	//endpoint customer
	mRoute.HandleFunc("/api/customer", getListUser).Methods("GET")
	mRoute.HandleFunc("/api/customer/{id}", getUser).Methods("GET")
	mRoute.HandleFunc("/api/customer", createUser).Methods("POST")
	mRoute.HandleFunc("/api/customer/{id}", updateUser).Methods("PUT")
	mRoute.HandleFunc("/api/customer/{id}", deleteUser).Methods("DELETE")
	//end endpoint customer

	log.Fatal(http.ListenAndServe(port, mRoute))
}
