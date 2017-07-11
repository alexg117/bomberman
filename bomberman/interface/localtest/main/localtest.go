package main

import (
  //"bufio"
  //"fmt"
  //"reflect"
  "log"
  //"bufio"
  //"math"
	//"math/rand"
  //"os"
  //"strconv"
  //"strings"
  //"time"
  //"bareBones"
  "bomberman/board"
  //"bomberman/message"
  //"bareBones"
  "bomberman/strategy"
)
func moveToCode(s string) int8 {
    switch s {
      case "mu":
        return 1
      case "mr":
        return 2
      case "md":
        return 3
      case "ml":
        return 4
      case "tu":
        return 5
      case "tr":
        return 6
      case "td":
        return 7
      case "tl":
        return 8
      case "b":
        return 9
      case "bp":
        return 10
      case "op":
        return 11
      case "buy_count":
        return 12
      case "buy_pierce":
        return 13
      case "buy_range":
        return 14
      case "buy_block":
        return 15
    }
    return 0 //No move
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
  h1 := harambe.Harambe{board.Board{},board.Pieces{},moveCodes,false,-1}
  h2 := harambe.Harambe{board.Board{},board.Pieces{},moveCodes,false,-1}
  board.DefaultBoard(&h1.CurrentBoard)
  board.DefaultBoard(&h2.CurrentBoard)
  board.DefaultPieces(&h1.CurrentBoard,&h1.CurrentPieces)
  h2.CurrentPieces=board.Pieces{}
  board.CopyPieces(&h2.CurrentPieces,&h1.CurrentPieces)
  b:=&h1.CurrentBoard
  p:=board.Pieces{}
  h2.CurrentBoard.PlayerIndex=true;
  board.CopyPieces(&p,&(h1.CurrentPieces))
  for x:=0;x<50; x++ {
    var nextMove int8
    log.Println("Turn",x)
    if ((p.MoveIterator&&p.MoveOrder[1])||((!p.MoveIterator&&p.MoveOrder[0]))) {
      log.Println("Harambe 2's turn")
      log.Println(p.O.BluePortal)
      log.Println(p.O.OrangePortal)
      nextMove=moveToCode(harambe.GetMove(&h2))
    } else {
      log.Println("Harambe 1's turn")
      log.Println(p.P.BluePortal)
      log.Println(p.P.OrangePortal)
      nextMove=moveToCode(harambe.GetMove(&h1))
    }
    log.Println("The move is",moveCodes[nextMove])
    if (!board.ApplyMove(b,&p,nextMove)) {
    if (nextMove!=9) {log.Println("ILLEGAL MOVE");break
    } else {
      board.IterateMove(b,&p,nextMove)
      }
    }
    //board.PrintBoard(b,&p,x)
    killed:=false
    if (p.Trails[p.P.X*11+p.P.Y]) {
      log.Println("Harambe 1 killed")
      killed=true
    }
    if (p.Trails[p.O.X*11+p.O.Y]) {
      log.Println("Harambe 2 killed")
      killed=true
    }
    if (killed) {break}
    board.CopyPieces(&h2.CurrentPieces,&p)
    board.CopyPieces(&h1.CurrentPieces,&p)
    board.CopyAvatar(&h2.CurrentPieces.P,&p.O)
    board.CopyAvatar(&h2.CurrentPieces.O,&p.P)
  }
}
