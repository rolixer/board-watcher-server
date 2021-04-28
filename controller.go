package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func main() {

	AddMove("a", Move{"Rook", "A1", "A5", false})

	go Start()
	fmt.Println("Server słucha na porcie 8181")

	for {
		var game string
		var move string
		fmt.Println("Podaj id gry: ")
		_, err := fmt.Scanf("%s\n", &game)

		if err != nil {
			log.Println(err)
		}

		fmt.Println("Podaj ruch w formacie {\"piece\":\"nazwa figury\",\"form\":\"pole poczatkowe\",\"to\":\"pole koncowe\",\"color\":true/false}")
		_, err = fmt.Scanf("%s\n", &move)

		if err != nil {
			log.Println(err)
		}

		var m Move

		err = json.Unmarshal([]byte(move), &m)
		if err != nil {
			log.Println(err)
			continue
		}

		AddMove(game, m)
	}
}
