package main

/*
 * Bomberman AI
 *
 * author: Matt Poegel
 */

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	//"time"
	"bomberman/strategy"
	"bomberman/message"
	"bomberman/board"
)

const serverURL = "http://aicomp.io/api/"

// Strategy defines approach for choosing the next move
type Strategy interface {
	Execute(*message.Message) string
}

// Player encapsulates the user's credentials
type Player struct {
	DevKey   string
	Username string
}


// NewPlayer reads the given credential file and returns a new Player
func NewPlayer(credentialFile *string) *Player {
	fp, err := os.Open(*credentialFile)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			panic(err)
		}
	}()

	buf := make([]byte, 1024)
	n, err := fp.Read(buf)
	if err != nil {
		panic(err)
	}
	if n == 0 {
		return nil
	}
	res := string(buf[:n])
	strres := string(res)
	bits := strings.Split(strres, "\n")
	p := &Player{bits[0], bits[1]}
	return p
}

func decodeResponse(resp *http.Response, msg *message.Message) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	dec := json.NewDecoder(strings.NewReader(string(body)))
	if err := dec.Decode(&msg); err != nil {
		log.Print(string(body))
		panic(err)
	}
}
// PracticeGame joins a practice game and begins making moves using the given strategy
func PracticeGame(player *Player, strat harambe.Harambe) {
	//resp, err := http.PostForm(serverURL+"games/search",
	resp, err := http.PostForm(serverURL+"games/search",
		url.Values{"devkey": {player.DevKey}, "username": {player.Username}})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var msg message.Message
	decodeResponse(resp,&msg)
	log.Printf("Joined game: %s\n", msg.GameID)
	log.Printf("PlayerID: %s\n", msg.PlayerID)
	log.Printf("Updating Board with new info")
	harambe.UpdateBoard(&strat,&msg)
	for {
		//How long may the strategy think?
		move:=harambe.GetMove(&strat)
		log.Printf("Move: %s\n", move)
		resp, err := http.PostForm(serverURL+"games/submit/"+msg.GameID,
			url.Values{"devkey": {player.DevKey}, "playerID": {msg.PlayerID}, "move": {move}})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	  decodeResponse(resp,&msg)
		log.Printf("Updating Board with new info")
		harambe.UpdateBoard(&strat,&msg)
		if msg.State == "complete" {
			break
		}
	}
	log.Printf("Game over\n")
}

func main() {
	moveCodes:=make([]string, 16)
	moveCodes[0]=""
	moveCodes[1]="mu"
	moveCodes[2]="mr"
	moveCodes[3]="md"
	moveCodes[4]="ml"
	moveCodes[5]="tu"
	moveCodes[6]="tr"
	moveCodes[7]="td"
	moveCodes[8]="tl"
	moveCodes[9]="b"
	moveCodes[10]="bp"
	moveCodes[11]="op"
	moveCodes[12]="buy_count"
	moveCodes[13]="buy_pierce"
	moveCodes[14]="buy_range"
	moveCodes[15]="buy_block"
  args := os.Args
  if len(args) < 1 {
      fmt.Fprintf(os.Stderr, "Usage: go run ai.go \n")
      os.Exit(1)
  }
  player := &Player{"5823c64504d7ce4a44c38a74",
		 "harambe"}
  log.Printf("Loaded credentials for: %s\n", player.Username)
  strat := harambe.Harambe{board.Board{},board.Pieces{},moveCodes,false,-1}
  log.Printf("Harambe initialized")

  PracticeGame(player, strat)
}
