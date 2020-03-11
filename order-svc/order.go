package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
)

const (
	//instance *Connection
	//NATSURL   = "nats://nats-streaming:4222"
	port      = ":8002"
	clusterID = "test-cluster"
	channel   = "order-notification"
	durableID = "restaurant-service-durable"
)

// type Connection struct {
// 	Nats *nats.Conn
// }

//models order
type Order struct {
	ID       string `json:"id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

//end models order

//models notif
type Notif struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

//end models notif

//init var as  a slice struct
var orders []Order
var notifs []Notif

//handlers order
func getListOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) //get param

	//Loop through books and find id
	for _, item := range orders {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Order{})
}

//asynchronus
func createOrderAsyn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order Order
	//var sc stan.Conn
	_ = json.NewDecoder(r.Body).Decode(&order)
	order.ID = strconv.Itoa(rand.Intn(10000000))
	orders = append(orders, order)
	json.NewEncoder(w).Encode(order)

	messgBuff, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshalling object.")
		fmt.Println(err)
	}

	var natsConn *nats.Conn
	natsConn, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	conn, err := stan.Connect(clusterID, "client-asynchronus", stan.NatsConn(natsConn))
	if err != nil {
		log.Print(err)
	}

	ackHandler := func(ackedNuid string, err error) {
		if err != nil {
			log.Printf("Error publishing message id %s: %v\n", ackedNuid, err.Error())
		} else {
			log.Printf("Received ACK for message id %s\n", ackedNuid)
		}
	}
	defer natsConn.Close()

	asynid, err := conn.PublishAsync(channel, messgBuff, ackHandler)
	if err != nil {
		log.Println("Error publish message!", asynid, err)
	}
	log.Printf("Published [%s] : '%s'\n", channel, messgBuff)

	defer conn.Close()
}

//synchronus
func createOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var order Order
	//var sc stan.Conn
	_ = json.NewDecoder(r.Body).Decode(&order)
	order.ID = strconv.Itoa(rand.Intn(10000000))
	orders = append(orders, order)
	json.NewEncoder(w).Encode(order)

	messgBuff, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Error marshalling object.")
		fmt.Println(err)
	}

	var natsConn *nats.Conn
	natsConn, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer natsConn.Close()

	conn, err := stan.Connect(clusterID, "client-order-2", stan.NatsConn(natsConn))
	if err != nil {
		log.Print(err)
	}

	defer conn.Close()

	//by default func conn.Publish is synchronus
	err = conn.Publish(channel, messgBuff)
	if err != nil {
		log.Println("Error publish message!")
	}
	log.Printf("Published [%s] : '%s'\n", channel, messgBuff)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range orders {
		if item.ID == params["id"] {
			orders = append(orders[:index], orders[index+1:]...)
			var order Order
			_ = json.NewDecoder(r.Body).Decode(&order)
			order.ID = params["id"]
			orders = append(orders, order)
			json.NewEncoder(w).Encode(order)
			return
		}
	}
	json.NewEncoder(w).Encode(orders)
}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range orders {
		if item.ID == params["id"] {
			orders = append(orders[:index], orders[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(orders)
}

//end handlers order

func main() {
	//initial data -implement database
	orders = append(orders, Order{ID: "1", Item: "Ayam", Quantity: 2})
	//end initial data

	//init router
	mRoute := mux.NewRouter()

	//endpoint order
	mRoute.HandleFunc("/api/order", getListOrder).Methods("GET")
	mRoute.HandleFunc("/api/order/{id}", getOrder).Methods("GET")
	mRoute.HandleFunc("/api/order", createOrder).Methods("POST")
	//asychronus nats
	mRoute.HandleFunc("/api/order/asyn", createOrderAsyn).Methods("POST")
	//end asynchronus nats
	mRoute.HandleFunc("/api/order/{id}", updateOrder).Methods("PUT")
	mRoute.HandleFunc("/api/order/{id}", deleteOrder).Methods("DELETE")
	//end endpoint order

	log.Println("run server order at port :8002")
	log.Fatal(http.ListenAndServe(port, mRoute))

}
