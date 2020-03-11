package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

const (
	port      = ":8001"
	clusterID = "test-cluster"
	clientID  = "restaurant-svc"
	channel   = "order-notification"
	durableID = "restaurant-service-durable"
)

//models restaurant
type Restaurant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

//end models restaurant

//models notif
type Notif struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

//end models notif

//init var as  a slice struct
var restaurants []Restaurant
var notifs []Notif

//handlers Restaurant
func getListRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(restaurants)
}

func getRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) //get param

	//Loop through restaurant and find id
	for _, item := range restaurants {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Restaurant{})
}

func createRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var restaurant Restaurant
	_ = json.NewDecoder(r.Body).Decode(&restaurant)
	restaurant.ID = strconv.Itoa(rand.Intn(10000000))
	restaurants = append(restaurants, restaurant)
	json.NewEncoder(w).Encode(restaurant)
}

func updateRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range restaurants {
		if item.ID == params["id"] {
			restaurants = append(restaurants[:index], restaurants[index+1:]...)
			var restaurant Restaurant
			_ = json.NewDecoder(r.Body).Decode(&restaurant)
			restaurant.ID = params["id"]
			restaurants = append(restaurants, restaurant)
			json.NewEncoder(w).Encode(restaurant)
			return
		}
	}
	json.NewEncoder(w).Encode(restaurants)
}

func deleteRestaurant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range restaurants {
		if item.ID == params["id"] {
			restaurants = append(restaurants[:index], restaurants[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(restaurants)
}

//end handlers restaurant

func main() {
	//initial data -implement database
	restaurants = append(restaurants, Restaurant{ID: "1", Name: "Ayam", Description: "ayam panggang dengan saos"})
	//end initial data

	var natsConn *nats.Conn
	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer natsConn.Close()

	conn, err := stan.Connect(clusterID, clientID, stan.NatsConn(natsConn))
	if err != nil {
		log.Print(err)
	}

	//initial message handler
	msgHandle := func(m *stan.Msg) {
		log.Print("Got new order! ", m.Subject)
		log.Print("Data order: ", string(m.Data))
	}
	conn.Subscribe(channel, msgHandle)
	defer conn.Close()

	//init router
	mRoute := mux.NewRouter()

	//endpoint book
	mRoute.HandleFunc("/api/restaurant", getListRestaurant).Methods("GET")
	mRoute.HandleFunc("/api/restaurant/{id}", getRestaurant).Methods("GET")
	mRoute.HandleFunc("/api/restaurant", createRestaurant).Methods("POST")
	mRoute.HandleFunc("/api/restaurant/{id}", updateRestaurant).Methods("PUT")
	mRoute.HandleFunc("/api/restaurant/{id}", deleteRestaurant).Methods("DELETE")
	//end endpoint book

	log.Println("Run server restaurant on port :8001")
	log.Fatal(http.ListenAndServe(port, mRoute))
}
