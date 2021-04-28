package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var pieces = map[string]int{
	"Rook":   2,
	"Knight": 3,
	"Bishop": 4,
	"Queen":  5,
	"King":   6,
	"Pawn":   1,
}

var fields = map[string]int{
	"a8": 0,
	"b8": 1,
	"c8": 2,
	"d8": 3,
	"e8": 4,
	"f8": 5,
	"g8": 6,
	"h8": 7,
	"a7": 8,
	"b7": 9,
	"c7": 10,
	"d7": 11,
	"e7": 12,
	"f7": 13,
	"g7": 14,
	"h7": 15,
	"a6": 16,
	"b6": 17,
	"c6": 18,
	"d6": 19,
	"e6": 20,
	"f6": 21,
	"g6": 22,
	"h6": 23,
	"a5": 24,
	"b5": 25,
	"c5": 26,
	"d5": 27,
	"e5": 28,
	"f5": 29,
	"g5": 30,
	"h5": 31,
	"a4": 32,
	"b4": 33,
	"c4": 34,
	"d4": 35,
	"e4": 36,
	"f4": 37,
	"g4": 38,
	"h4": 39,
	"a3": 40,
	"b3": 41,
	"c3": 42,
	"d3": 43,
	"e3": 44,
	"f3": 45,
	"g3": 46,
	"h3": 47,
	"a2": 48,
	"b2": 49,
	"c2": 50,
	"d2": 51,
	"e2": 52,
	"f2": 53,
	"g2": 54,
	"h2": 55,
	"a1": 56,
	"b1": 57,
	"c1": 58,
	"d1": 59,
	"e1": 60,
	"f1": 61,
	"g1": 62,
	"h1": 63,
}

type Move struct {
	Piece string `json:"piece"`
	From  string `json:"from"`
	To    string `json:"to"`
	Color bool   `json:"color"`
}

type game struct {
	moves       []Move
	connections []*websocket.Conn
}

type Games map[string]*game

var upgrader = websocket.Upgrader{} // use default options
var games = make(Games)

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	log.Print(r.URL.Query().Get("id"))

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func run() {

	g1 := &game{[]Move{}, []*websocket.Conn{}}

	games["xD"] = g1

	if _, ok := games["t"]; !ok {
		log.Printf("Game not found")
		return
	}

}

func addMoveReq(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	var m []Move

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &m)

	if err != nil {
		log.Fatal(err)
	}

	if games[id] == nil {
		g := &game{[]Move{}, []*websocket.Conn{}}
		games[id] = g
	}

	games[id].moves = append(games[id].moves, m[0])
	sendMove(id, m[0])

}

func AddMove(id string, m Move) {

	if games[id] == nil {
		g := &game{[]Move{}, []*websocket.Conn{}}
		games[id] = g
	}

	games[id].moves = append(games[id].moves, m)
	sendMove(id, m)

}

func sendGame(id string, c *websocket.Conn) {
	for _, m := range games[id].moves {
		moveStr := fmt.Sprintf("{%s: [0, false], %s:[%s, %t]}", m.From, m.To, m.Piece, m.Color)

		c.WriteMessage(websocket.TextMessage, []byte(moveStr))
	}

}

func sendMove(id string, m Move) {

	for _, conn := range games[id].connections {
		moveStr := fmt.Sprintf("{%s: [0, false], %s:[%s, %t]}", m.From, m.To, m.Piece, m.Color)

		conn.WriteMessage(websocket.TextMessage, []byte(moveStr))

	}

}

func watch(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if games[id] == nil {
		fmt.Fprintf(w, "Brak gry o takim id")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	games[id].connections = append(games[id].connections, conn)
	sendGame(id, conn)

	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
	}
}

func Start() {
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	http.HandleFunc("/move", addMoveReq)
	http.HandleFunc("/watch", watch)
	log.Fatal(http.ListenAndServe("localhost:8181", nil))
}
