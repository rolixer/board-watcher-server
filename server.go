package boardwatcherserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var pieces = map[string]int{
	"ROOK":   2,
	"KNIGHT": 3,
	"BISHOP": 4,
	"QUEEN":  5,
	"KING":   6,
	"PAWN":   1,
}

var fields = map[string]int{
	"A8": 0,
	"B8": 1,
	"C8": 2,
	"D8": 3,
	"E8": 4,
	"F8": 5,
	"G8": 6,
	"H8": 7,
	"A7": 8,
	"B7": 9,
	"C7": 10,
	"D7": 11,
	"E7": 12,
	"F7": 13,
	"G7": 14,
	"H7": 15,
	"A6": 16,
	"B6": 17,
	"C6": 18,
	"D6": 19,
	"E6": 20,
	"F6": 21,
	"G6": 22,
	"H6": 23,
	"A5": 24,
	"B5": 25,
	"C5": 26,
	"D5": 27,
	"E5": 28,
	"F5": 29,
	"G5": 30,
	"H5": 31,
	"A4": 32,
	"B4": 33,
	"C4": 34,
	"D4": 35,
	"E4": 36,
	"F4": 37,
	"G4": 38,
	"H4": 39,
	"A3": 40,
	"B3": 41,
	"C3": 42,
	"D3": 43,
	"E3": 44,
	"F3": 45,
	"G3": 46,
	"H3": 47,
	"A2": 48,
	"B2": 49,
	"C2": 50,
	"D2": 51,
	"E2": 52,
	"F2": 53,
	"G2": 54,
	"H2": 55,
	"A1": 56,
	"B1": 57,
	"C1": 58,
	"D1": 59,
	"E1": 60,
	"F1": 61,
	"G1": 62,
	"H1": 63,
}

// Move struct storing a single move
type Move struct {
	Piece string `json:"piece"`
	From  string `json:"from"`
	To    string `json:"to"`
	Color bool   `json:"color"`
}

// Game struct storing a single game
type Game struct {
	Moves       []Move
	Connections []*websocket.Conn
}

// Games map storing games associeted with id's
type Games map[string]*Game

var upgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
var games = make(Games)
var gamesAI = make(map[string]string)

// Home handler, sends simple text response for testing purposes
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}

func run() {

	g1 := &Game{[]Move{}, []*websocket.Conn{}}

	games["xD"] = g1

	if _, ok := games["t"]; !ok {
		log.Printf("Game not found")
		return
	}

}

// AddMoveReq handler, add move to a game
func AddMoveReq(w http.ResponseWriter, r *http.Request) {

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
		g := &Game{[]Move{}, []*websocket.Conn{}}
		games[id] = g
	}

	for _, m := range m {
		games[id].Moves = append(games[id].Moves, m)
		SendMove(id, m)
	}

}

// AddMove ads move to a game
func AddMove(id string, m Move) {

	if games[id] == nil {
		g := &Game{[]Move{}, []*websocket.Conn{}}
		games[id] = g
	}

	games[id].Moves = append(games[id].Moves, m)
	SendMove(id, m)

}

// RecertMove handler, reverts last move in game
func RevertMove(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if games[id] == nil {
		w.Write([]byte("Brak gry o podanym id"))
		return
	}

	moves_r := games[id].Moves
	moves_r = moves_r[:len(moves_r)-1]
	games[id].Moves = moves_r

	for _, c := range games[id].Connections {
		c.WriteMessage(websocket.TextMessage, []byte("REVERT"))
	}

	w.Write([]byte("Cofnięto"))
}

//SendGame sends whole game to newly established connection
func SendGame(id string, c *websocket.Conn) {
	for _, m := range games[id].Moves {
		moveStr := fmt.Sprintf("{\"piece\":%d, \"from\":%d, \"to\":%d, \"color\":%t}", pieces[m.Piece], fields[m.From], fields[m.To], m.Color)

		c.WriteMessage(websocket.TextMessage, []byte(moveStr))
	}

}

// SendMove sends move to every client watching game witch given id
func SendMove(id string, m Move) {

	for _, conn := range games[id].Connections {
		moveStr := fmt.Sprintf("{\"piece\":%d, \"from\":%d, \"to\":%d, \"color\":%t}", pieces[m.Piece], fields[m.From], fields[m.To], m.Color)

		conn.WriteMessage(websocket.TextMessage, []byte(moveStr))

	}

}

// Watch establishesh websocket connection to watch live game if game exists
func Watch(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if games[id] == nil {
		w.Write([]byte("Nie znaleźono gry o podanym ID"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	games[id].Connections = append(games[id].Connections, conn)
	SendGame(id, conn)

	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived:
					return
				}
			}
			log.Println("Error during message reading:", err)
			continue
		}
	}
}

// StartGame handler registering new game
func StartGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if games[id] == nil {
		g := &Game{[]Move{}, []*websocket.Conn{}}
		games[id] = g
	}

	w.Write([]byte("Rozpoczęto rozgrywkę"))
}

// SaveAI handler saving AI analyzed position
func SaveAI(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal(err)
	}

	gamesAI[id] = string(body)

}

// GetAI handler downloads AI analyzed position
func GetAI(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if _, ok := gamesAI[id]; ok {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(gamesAI[id]))
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte("Brak pozycji o takim ID"))
	}
}

// Start the server
func Start() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/move", AddMoveReq)
	http.HandleFunc("/watch", Watch)
	http.HandleFunc("/revert", RevertMove)
	http.HandleFunc("/start", StartGame)
	http.HandleFunc("/AI/add", SaveAI)
	http.HandleFunc("/AI/get", GetAI)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
