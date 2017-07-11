package harambe

import (
  "bufio"
  //"fmt"
  //"reflect"
  "log"
  //"bufio"
  //"math"
	"math/rand"
  "os"
  "strconv"
  "strings"
  "time"
  //"bareBones"
  "bomberman/board"
  "bomberman/message"
  "bareBones"
)


type Harambe struct {
  CurrentBoard   board.Board
  CurrentPieces  board.Pieces
  MoveList       []string
  BoughtCount    bool
  Apocalypse     int
}
type SquareData struct {
  owner            *board.Avatar
  dist             int8
  pbPortal         *board.Portal
  poPortal         *board.Portal
  obPortal         *board.Portal
  ooPortal         *board.Portal
  facing           int8
  src              int8
  unsafe           []int8
}
type Edge struct {
  dest          int8
  weight        int8
  moves         *[]int8
}
type SDGraphHolder struct {
  sd            *[]SquareData
  e             [][]Edge
}
func manhattanDist(p1 *board.Avatar,p2 *board.Avatar) int8 {
  dx:=p1.X-p2.X
  if dx<0 {dx=-dx}
  dy:=p1.Y-p2.Y
  if dy<0 {dy=-dy}
  return dy+dx
}
func manhattanDistInt(i1 int8,i2 int8) int8 {
  dx:=i2/11-i1/11
  if dx<0 {dx=-dx}
  dy:=i2%11-i1%11
  if dy<0 {dy=-dy}
  return dy+dx
}
func farthestPath(paths *[][]int8,values *[]int) {
  maxDist:=int8(0)
  maxIndex:=0
  var dist int8
  for x:=0;x<len(*paths);x++ {
    dist=manhattanDistInt((*paths)[x][0],(*paths)[x][len((*paths)[x])-1])
    if ((dist>maxDist)&&((*values)[x]>0)) {
      maxIndex=x
      maxDist=dist
    }
  }
  (*values)[maxIndex]=100000000
}
//Harmonic simultaneous 4-way blockage detector
func freeSquares(b *board.Board,p *board.Pieces,a *board.Avatar) int8 {
  playerLoc:=a.X*11+a.Y
  freeSquares:=int8(4)
  distances:=[]int8{-1,-11,1,11}
  for x:=0;x<4;x++ {
    if ((b.HardBlockBoard[playerLoc+distances[x]])||
    (p.SoftBlockBoard[playerLoc+distances[x]])) {
      freeSquares--
    }
  }
  return freeSquares
}
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
func copySquareData(dest *SquareData,src SquareData) {
  dest.owner=src.owner
  dest.dist=src.dist
  dest.pbPortal=src.pbPortal
  dest.poPortal=src.poPortal
  dest.obPortal=src.obPortal
  dest.ooPortal=src.ooPortal
  dest.facing=src.facing
  dest.src=src.src
  dest.unsafe=make([]int8,len(src.unsafe))
  copy(dest.unsafe,src.unsafe)
}
func apocalypseTrails(i int) []bool {
  apoBools:=make([]bool,121)
  if (i<=0) {return apoBools}
  apo:=[]int{11,10,9 ,8 ,7 ,6 ,5 ,4 ,3 ,2 ,1 ,
            12,29,28,27,26,25,24,23,22,21,20,
            13,30,43,42,41,40,39,38,37,36,19,
            14,31,44,53,52,51,50,49,48,35,18,
            15,32,45,54,59,58,57,56,47,34,17,
            16,33,46,55,60,61,60,55,46,33,16,
            17,34,47,56,57,58,59,54,45,32,15,
            18,35,48,49,50,51,52,53,44,31,14,
            19,36,37,38,39,40,41,42,43,30,13,
            20,21,22,23,24,25,26,27,28,29,12,
            1 ,2 ,3 ,4 ,5 ,6 ,7 ,8 ,9 ,10,11}
  for x:=0;x<121;x++ {
    if (i>=apo[x]) {apoBools[x]=true}
  }
  return apoBools
}
func getIntPair(s []string) [2]int {
  i1,err:=strconv.Atoi(s[0])
  if err != nil {
    panic(err)
  }
  i2,err:=strconv.Atoi(s[1])
  if err != nil {
    panic(err)
  }
  return [2]int{i1,i2}
}
func parsePortal(p* board.Portal,portal map[string]interface{}) {
  (*p).X=int8(portal["x"].(float64))
  (*p).Y=int8(portal["y"].(float64))
  (*p).Orientation=int8(portal["direction"].(float64))
}
func parseAvatar(a* board.Avatar,m map[string]interface{}) {
  (*a).X=int8(m["x"].(float64))
  (*a).Y=int8(m["y"].(float64))
  (*a).Count=int8(m["bombCount"].(float64))
  (*a).Pierce=int8(m["bombPierce"].(float64))
  (*a).BombRange=int8(m["bombRange"].(float64))
  (*a).Coins=int16(m["coins"].(float64))
  (*a).Orientation=int8(m["orientation"].(float64))
  (*a).OrangePortal=&board.Portal{}
  if (m["orangePortal"])!=nil {
    parsePortal((*a).OrangePortal,m["orangePortal"].(map[string]interface {}))
  } else {(*a).OrangePortal=nil}
  (*a).BluePortal=&board.Portal{}
  if (m["bluePortal"])!=nil {
    parsePortal((*a).BluePortal,m["bluePortal"].(map[string]interface {}))
  } else {(*a).BluePortal=nil}
}
func UpdateBoard(h *Harambe, msg *message.Message) {
  h.CurrentPieces=board.Pieces{}
  if (h.CurrentBoard.BoardSize==0) {board.DefaultBoard(&h.CurrentBoard)}
  (*h).CurrentPieces.SoftBlockBoard=make([]bool, len((*msg).SoftBlockBoard))
  for i:=int8(0);i<(*h).CurrentBoard.BoardSize*(*h).CurrentBoard.BoardSize;i++ {
    (*h).CurrentPieces.SoftBlockBoard[i]=(((*msg).SoftBlockBoard)[i]==1)
  }
  (*h).CurrentPieces.MoveIterator=((*msg).MoveIterator==1)
  h.CurrentPieces.MoveOrder=make([]bool,2)
  for i:=0;i<2;i++ {h.CurrentPieces.MoveOrder[i]=((*msg).MoveOrder[i]==1)}
  if (h.CurrentPieces.MoveIterator) {h.CurrentBoard.PlayerIndex=h.CurrentPieces.MoveOrder[1]} else
  {h.CurrentBoard.PlayerIndex=h.CurrentPieces.MoveOrder[0]}
  (*h).CurrentPieces.Bombs=nil
  m,ok:=(*msg).BombMap.(map[string]interface{})
  if (ok) {
    for k, v := range m {
      m2,ok:=v.(map[string]interface{})
      if (ok) {//Parse bomb
        i:=getIntPair(strings.Split(k,","))
        (*h).CurrentPieces.Bombs=append((*h).CurrentPieces.Bombs,board.Bomb{int8(i[0]),int8(i[1]),(m2["owner"]==0),int8(m2["tick"].(float64))})
      }
    }
  }
  (*h).CurrentPieces.Trails=make([]bool,(*h).CurrentBoard.BoardSize*(*h).CurrentBoard.BoardSize)
  m,ok=(*msg).TrailMap.(map[string]interface{})
  if (ok) {
    for k := range m {
      i:=getIntPair(strings.Split(k,","))
      (*h).CurrentPieces.Trails[int8(i[0])*(*h).CurrentBoard.BoardSize+int8(i[1])]=true
    }
  }
  (*h).CurrentPieces.P=board.Avatar{}
  m,ok=(*msg).Player.(map[string]interface{})
  if (ok) {
    parseAvatar(&((*h).CurrentPieces.P),m)
  }
  (*h).CurrentPieces.O=board.Avatar{}
  m,ok=(*msg).Opponent.(map[string]interface{})
  if (ok) {
    parseAvatar(&((*h).CurrentPieces.O),m)
  }
}
//Checks if we can capture the square
func canWeCapture(b *board.Board,p *board.Pieces,maxDist int8,reachableSquares *[]SquareData,nextSquare int8,destSquare int8,nextDist int8,advantage float32) bool {
  claimable:=false
  if (((*reachableSquares)[destSquare].owner==nil)) {
    //It's unowned. Can we claim it within the time limit?
    if (nextDist<maxDist)||((nextDist<=maxDist)&&((!p.MoveIterator)||
    (((*reachableSquares)[nextSquare].owner==&p.P)&&p.MoveOrder[1]==b.PlayerIndex)||
    (((*reachableSquares)[nextSquare].owner==&p.O)&&p.MoveOrder[1]!=b.PlayerIndex))) {claimable=true}
  } else if ((*reachableSquares)[destSquare].owner==(*reachableSquares)[nextSquare].owner) {
    //We own it. Can we claim it faster?
    if ((*reachableSquares)[destSquare].dist>nextDist) {claimable=true}
  } else {
    ourTime:=nextDist
    theirTime:=(*reachableSquares)[destSquare].dist
    if theirTime==0 {return false} //Can't claim starting square
    for ((ourTime>2)&&(theirTime>2)) {
      ourTime-=2
      theirTime-=2
    }
    if ourTime>2 {
      return false
    }
    if theirTime>2 {
      return true
    }
    //They own it. Can we seize it?
    priority:=advantage
    //First see if we are the owner. If we are not, we must flip priority.
    if (*reachableSquares)[nextSquare].owner==&p.O {
      priority=-priority
    }
    //Now we can see if we can take the square
    if (priority==1.5) {return true}
    if (priority==-1.5) {return true}
    if (priority==0.5) {
      if (ourTime==1) {return true}
      return false
    }
    if (priority==-0.5) {
      if (theirTime==1) {return false}
      return true
    }
    if (float32(nextDist)-priority< float32((*reachableSquares)[destSquare].dist)) {
      //We can seize this square
      claimable=true
    }
  } //end else
  return claimable
}
func getReachableSquares(b* board.Board,p *board.Pieces,maxDist int8,reachableSquares *[]SquareData)  {
  numSquares:=b.BoardSize*b.BoardSize
  var nextPortal *board.Portal
  var nextDestPortal *board.Portal
  var nextSolidBlock int8
  var nextOrientation int8
  (*reachableSquares)=make([]SquareData,numSquares)
  for i:=0;i<121;i++ {
    (*reachableSquares)[i]=SquareData{}
    (*reachableSquares)[i].unsafe=make([]int8,10)
  }
  (*reachableSquares)[p.P.X*b.BoardSize+p.P.Y]=SquareData{&p.P,0,p.P.BluePortal,p.P.OrangePortal,p.O.BluePortal,p.O.OrangePortal,p.P.Orientation,0,make([]int8,10)}
  (*reachableSquares)[p.O.X*b.BoardSize+p.O.Y]=SquareData{&p.O,0,p.P.BluePortal,p.P.OrangePortal,p.O.BluePortal,p.O.OrangePortal,p.O.Orientation,0,make([]int8,10)}
  var advantage float32
  if (p.MoveIterator) {
    if (p.MoveOrder[1]!=b.PlayerIndex) {
      advantage=-1.5
    } else {
      advantage=1.5
    }
  } else {
    advantage=0.5
    if (p.MoveOrder[0]!=b.PlayerIndex) {advantage=-0.5}
  }

  var int8Queue queue.Queue
  var p1square interface{}
  var p2square interface{}
  p1square=int8(p.P.X*b.BoardSize+p.P.Y)
  p2square=int8(p.O.X*b.BoardSize+p.O.Y)
  queue.AddToQueue(&int8Queue,&p1square)
  queue.AddToQueue(&int8Queue,&p2square)
  //10 rounds
  nextEntry:=queue.GetFromQueue(&int8Queue)
  var nextSquare int8
  for nextEntry!=nil {
    nextSquare=(*nextEntry).(int8)
    adjacentSquares:=[]int8{nextSquare-11,nextSquare-1,nextSquare+11,nextSquare+1}
    turnsForPortal:=int8(99)
    for i:=int8(0);i<4;i++ {
      //First we make sure the square is free
      if ((b.HardBlockBoard[adjacentSquares[i]])||
      (p.SoftBlockBoard[adjacentSquares[i]])) {
        //The square's not free, maybe we can use it for a portal
        if (turnsForPortal>2)&&(b.HardBlockBoard[adjacentSquares[i]]) {turnsForPortal=2;nextSolidBlock=adjacentSquares[i];nextOrientation=i}
        if ((turnsForPortal>1&&(*reachableSquares)[nextSquare].facing==i)&&(b.HardBlockBoard[adjacentSquares[i]])) {turnsForPortal=1;nextSolidBlock=adjacentSquares[i];nextOrientation=i}
        if (turnsForPortal>0) {
          if ((*reachableSquares)[nextSquare].owner==&p.P) {
            if (*reachableSquares)[nextSquare].pbPortal!=nil&&((*reachableSquares)[nextSquare].pbPortal.X==adjacentSquares[i]/11&&
              (*reachableSquares)[nextSquare].pbPortal.Y==adjacentSquares[i]%11)&&
              ((*reachableSquares)[nextSquare].pbPortal.Orientation+2)%4==i {
                turnsForPortal=0
                nextPortal=(*reachableSquares)[nextSquare].pbPortal
                nextDestPortal=(*reachableSquares)[nextSquare].poPortal
                nextSolidBlock=adjacentSquares[i]
                nextOrientation=i
              }
            if (*reachableSquares)[nextSquare].poPortal!=nil&&((*reachableSquares)[nextSquare].poPortal.X==adjacentSquares[i]/11&&
              (*reachableSquares)[nextSquare].poPortal.Y==adjacentSquares[i]%11)&&
              ((*reachableSquares)[nextSquare].poPortal.Orientation+2)%4==i {
                turnsForPortal=0
                nextPortal=(*reachableSquares)[nextSquare].poPortal
                nextDestPortal=(*reachableSquares)[nextSquare].pbPortal
                nextSolidBlock=adjacentSquares[i]
                nextOrientation=i
              }
          } else {
            if (*reachableSquares)[nextSquare].obPortal!=nil&&((*reachableSquares)[nextSquare].obPortal.X==adjacentSquares[i]/11&&
              (*reachableSquares)[nextSquare].obPortal.Y==adjacentSquares[i]%11)&&
              ((*reachableSquares)[nextSquare].obPortal.Orientation+2)%4==i {
                turnsForPortal=0
                nextPortal=(*reachableSquares)[nextSquare].obPortal
                nextDestPortal=(*reachableSquares)[nextSquare].ooPortal
                nextSolidBlock=adjacentSquares[i]
                nextOrientation=i
              }
            if (*reachableSquares)[nextSquare].ooPortal!=nil&&((*reachableSquares)[nextSquare].ooPortal.X==adjacentSquares[i]/11&&
              (*reachableSquares)[nextSquare].ooPortal.Y==adjacentSquares[i]%11)&&
              ((*reachableSquares)[nextSquare].ooPortal.Orientation+2)%4==i {
              turnsForPortal=0
              nextPortal=(*reachableSquares)[nextSquare].ooPortal
              nextDestPortal=(*reachableSquares)[nextSquare].obPortal
              nextSolidBlock=adjacentSquares[i]
              nextOrientation=i
            }
          }
        }
      } else {
        if !((board.BombAt(p,adjacentSquares[i]/11,adjacentSquares[i]%11)!=nil)) {
          //There is nothing outright blocking the square
          //See if we can claim it
          nextDist:=(*reachableSquares)[nextSquare].dist+1
          if ((*reachableSquares)[nextSquare].dist==0&&p.Trails[adjacentSquares[i]]) {nextDist++}
          if (canWeCapture(b,p,maxDist,reachableSquares,nextSquare,adjacentSquares[i],nextDist,advantage)) {
            (*reachableSquares)[adjacentSquares[i]]=SquareData{
              (*reachableSquares)[nextSquare].owner,
              nextDist,
              (*reachableSquares)[nextSquare].pbPortal,
              (*reachableSquares)[nextSquare].poPortal,
              (*reachableSquares)[nextSquare].obPortal,
              (*reachableSquares)[nextSquare].ooPortal,
              i,
              nextSquare,make([]int8,10)}
            if (*reachableSquares)[nextSquare].dist+1<maxDist {
              var addEntry interface{}
              addEntry=int8(adjacentSquares[i])
              queue.AddToQueue(&int8Queue,&addEntry)
            }
          }
        }
      }
    }
    //Now we tested adjacent squares
    //Try seeing where we can go by portal
    if (turnsForPortal<90) {
      //We are standing next to at least one solid block, see if we can portal through it
      var destPortals []*board.Portal
      var adjPortal *board.Portal
      if (turnsForPortal==0) {
        //We are standing next to a portal, see what we can do with that
        adjPortal=nextPortal
        destPortals=[]*board.Portal{nextDestPortal}
      } else {
        //See what we can do with either of our portals that might be out there.
        adjPortal=&board.Portal{nextSolidBlock/11,nextSolidBlock%11,(nextOrientation+2)%4}
        destPortals=make([]*board.Portal,2)
        if ((*reachableSquares)[nextSquare].owner==&p.P) {
          destPortals[0]=(*reachableSquares)[nextSquare].pbPortal
          destPortals[1]=(*reachableSquares)[nextSquare].poPortal
        } else {
          destPortals[0]=(*reachableSquares)[nextSquare].obPortal
          destPortals[1]=(*reachableSquares)[nextSquare].ooPortal
        }
      }
      //We shoot a portal by us if needed, and see if it leads us to our other portal.
      for j:=0;j<2;j++ {
        if (len(destPortals)<=j)||((destPortals[j])==nil) {continue}
        destSquare:=int8(destPortals[j].X*11+destPortals[j].Y)
        switch destPortals[j].Orientation {
        case 0:
          destSquare-=11
        case 1:
          destSquare--
        case 2:
          destSquare+=11
        case 3:
          destSquare++
        }
        //We made another portal. Traverse it and try to claim the square.
        //First we test to see if we can actually claim the square
        //Make sure it's not blocked
        if ((b.HardBlockBoard[destSquare])||
        (p.SoftBlockBoard[destSquare]))||
        (board.BombAt(p,destSquare/11,destSquare%11)!=nil)  {continue}
        nextDist:=int8((*reachableSquares)[nextSquare].dist+turnsForPortal+1)
        if ((*reachableSquares)[nextSquare].dist==0&&p.Trails[destSquare]&&turnsForPortal==0) {nextDist++}
        if (!canWeCapture(b,p,maxDist,reachableSquares,nextSquare,destSquare,nextDist,advantage)) {continue}
        var PnextBlue *board.Portal
        var PnextOrange *board.Portal
        var OnextBlue *board.Portal
        var OnextOrange *board.Portal
        if ((*reachableSquares)[nextSquare].owner==&p.P) {
          //Player is moving, enemy portals unchanged
          OnextBlue=(*reachableSquares)[nextSquare].obPortal
          OnextOrange=(*reachableSquares)[nextSquare].ooPortal
          if (j==0) {
            //We travel through our blue portal to the orange portal.
            PnextBlue=adjPortal
            PnextOrange=destPortals[j]
          } else {
          //We travel through our orange portal to the blue portal
          PnextBlue=destPortals[j]
          PnextOrange=adjPortal
          }
        } else {
          //Enemy is moving, player portals unchanged
          PnextBlue=(*reachableSquares)[nextSquare].pbPortal
          PnextOrange=(*reachableSquares)[nextSquare].poPortal
          if (j==0) {
            //We travel through our blue portal to the orange portal.
            OnextBlue=adjPortal
            OnextOrange=destPortals[j]
          } else {
            //We travel through our orange portal to the blue portal
            OnextBlue=destPortals[j]
            OnextOrange=adjPortal
          }
        }
        //Everything is set up and we know we can claim it
        (*reachableSquares)[destSquare]=SquareData{
          (*reachableSquares)[nextSquare].owner,
          nextDist,
          PnextBlue,
          PnextOrange,
          OnextBlue,
          OnextOrange,
          destPortals[j].Orientation,
          nextSquare,make([]int8,10)}
        if (nextDist<maxDist) {
          var addEntry interface{}
          addEntry=int8(destSquare)
          queue.AddToQueue(&int8Queue,&addEntry)
        }
      }
      //What happens if we shoot another portal down each of the directions?
      for k:=int8(0);k<4;k++ {
        //First find out where the portal will hit
        portalSquare:=nextSquare
        for (!b.HardBlockBoard[portalSquare])&&!p.SoftBlockBoard[portalSquare] {
          switch k {
          case 0:
            portalSquare-=11
          case 1:
            portalSquare--
          case 2:
            portalSquare+=11
          case 3:
            portalSquare++
          }
        }
        //We have the square. Now see where we can go
        newDestPortal:=&board.Portal{portalSquare/11,portalSquare%11,(k+2)%4}

        destSquare:=portalSquare
        switch newDestPortal.Orientation {
        case 0:
          destSquare-=11
        case 1:
          destSquare--
        case 2:
          destSquare+=11
        case 3:
          destSquare++
        }
        //Make sure we're not just ending up on the same square
        if (destSquare==nextSquare) {continue}
        //Make sure we're not blocked on portal exit
        if ((b.HardBlockBoard[destSquare])||
        (p.SoftBlockBoard[destSquare]))||
        (board.BombAt(p,destSquare/11,destSquare%11)!=nil)  {continue}
        //We need turnsForPortal to place the local portal plus one turn to fire the away portal plus one turn to traverse the portal
        nextDist:=int8((*reachableSquares)[nextSquare].dist+turnsForPortal+2)
        //If we're not facing the direction we want to shoot, turning is an additional move.
        if ((*reachableSquares)[nextSquare].facing!=k) {nextDist++}
        if (!canWeCapture(b,p,maxDist,reachableSquares,nextSquare,destSquare,nextDist,advantage)) {continue}
        //We now know that we can capture the tile. Make it so.
        var PnextBlue *board.Portal
        var PnextOrange *board.Portal
        var OnextBlue *board.Portal
        var OnextOrange *board.Portal
        if ((*reachableSquares)[nextSquare].owner==&p.P) {
          //Player is moving, enemy portals unchanged
          OnextBlue=(*reachableSquares)[nextSquare].obPortal
          OnextOrange=(*reachableSquares)[nextSquare].ooPortal
          //THIS SEGMENT MAY LABEL PORTAL COLORS INCORRECTLY
          if (k==0) {
          //We travel through our orange portal to the blue portal
            PnextBlue=newDestPortal
            PnextOrange=adjPortal
          } else {
            //We travel through our blue portal to the orange portal.
            PnextBlue=adjPortal
            PnextOrange=newDestPortal
          }
        } else {
          //Enemy is moving, player portals unchanged
          PnextBlue=(*reachableSquares)[nextSquare].pbPortal
          PnextOrange=(*reachableSquares)[nextSquare].poPortal
          if (k==0) {
          //We travel through our orange portal to the blue portal
            OnextBlue=newDestPortal
            OnextOrange=adjPortal
          } else {
            //We travel through our blue portal to the orange portal.
            OnextBlue=adjPortal
            OnextOrange=newDestPortal
          }
        }

        //Everything is set up and we know we can claim it
        (*reachableSquares)[destSquare]=SquareData{
          (*reachableSquares)[nextSquare].owner,
          nextDist,
          PnextBlue,
          PnextOrange,
          OnextBlue,
          OnextOrange,
          newDestPortal.Orientation,
          nextSquare,make([]int8,10)}
        if (nextDist<maxDist) {
          var addEntry interface{}
          addEntry=int8(destSquare)
          queue.AddToQueue(&int8Queue,&addEntry)
        }
      }
    }
    nextEntry=queue.GetFromQueue(&int8Queue)
  }
}
func getPortalSquare(b *board.Board,p *board.Pieces,startSquare int8,orientation int8) int8 {
  portalSquare:=startSquare
  for (!b.HardBlockBoard[portalSquare])&&!p.SoftBlockBoard[portalSquare] {
    switch orientation {
    case 0:
      portalSquare-=11
    case 1:
      portalSquare--
    case 2:
      portalSquare+=11
    case 3:
      portalSquare++
    }
  }
  return portalSquare
}
//Approximate which of the reachable squares are safe. May mark more squares than
//necessary as unsafe.
func mark4Death(b* board.Board,p *board.Pieces,reachableSquares *[]SquareData,apo int) {
  if (len(p.Bombs)==0) {return}
  pisces:=board.Pieces{}
  for i:=int8(0);i<6;i++ {
    apoBools:=apocalypseTrails(apo+int(i))
    if (apo>0) {
      for y:=0;y<121;y++ {
        if (apoBools[y]) {(*reachableSquares)[y].unsafe[i]=3}
      }
    }
    board.CopyPieces(&pisces,p)
    pisces.O.BombRange+=i
    pisces.O.Pierce+=i
    if (i>0) {pisces.Trails=make([]bool,121)}
    activeBombs:=[]board.Bomb{}
    i2:=int8(0);
    for i2< int8(len(pisces.Bombs)) {
      if ((pisces.Bombs[i2].Ticks)<i) {
        pisces.Bombs=append(pisces.Bombs[:i2], pisces.Bombs[i2+1:]...)
      } else {
        if ((pisces.Bombs[i2].Ticks)==i) {
          activeBombs=append(activeBombs,pisces.Bombs[i2])
        }
        i2++
      }
    }
    d1:=make([]bool,121)
    d2:=make([]bool,121)
    //Most basic case. Bombs exploding as is.
    for k3:=len(activeBombs)-1;k3>=0 ;k3-- {
      board.ProcessExplosion(b,&pisces,&activeBombs[k3],&d1,&d2)
    }
    //Detonate all bombs touching apocalypse Trails
    if (apo>0) {
      for trailCount:=int8(0);trailCount<121;trailCount++ {
        if (apoBools[trailCount]) {
          tempBomb:=board.BombAt(&pisces,trailCount/11,trailCount%11)
          if tempBomb!=nil {
            board.ProcessExplosion(b,&pisces,tempBomb,&d1,&d2)
          }
        }
      }
    }
    var p11 queue.Queue
    var p21 queue.Queue
    var b1 queue.Queue
    var b2 queue.Queue
    safeBombs:=0
    dangerBombs:=0
    safePortals:=0
    dangerPortals:=0
    for i2=int8(0);i2<121;i2++ {
      if (pisces.Trails[i2]) {
        (*reachableSquares)[i2].unsafe[i]=3 //Mark it as 3, definitely unsafe
        if ((*reachableSquares)[i2].owner==&p.O) {
          //First check for surprise chain bombs
          if (((*reachableSquares)[i2].dist<i-2)&&pisces.Trails[i2]) {
            //Opponent MAY be able to safely place a bomb here.
            safeBombs++
            var nextBomb interface{}
            nextBomb=board.Bomb{int8(i2/11),int8(i2%11),false,5}
            queue.AddToQueue(&b2,&nextBomb)
          }
          if (((*reachableSquares)[i2].dist<i)&&pisces.Trails[i2]) {
            //Opponent can put a bomb here but not without killing themselves
            dangerBombs++
            var nextBomb interface{}
            nextBomb=board.Bomb{int8(i2/11),int8(i2%11),false,5}
            queue.AddToQueue(&b1,&nextBomb)

          }
          if (((*reachableSquares)[i2].dist<i)) {
          //Opponent can only shoot a portal in the direction they are facing.
          //Doing so means likely death if an explosion goes through it.
            dangerPortals++
            nextPortal:=getPortalSquare(b,p,int8(i2),(*reachableSquares)[i2].facing)
            var Portalz interface{}
            Portalz=board.Portal{int8(nextPortal/11),int8(nextPortal%11),int8(((*reachableSquares)[i2].facing+2)%4)}
            queue.AddToQueue(&p11,&Portalz)
          }
          if (((*reachableSquares)[i2].dist<i-1)) {
            //Opponent can shoot a portal in any direction.
            //Doing so means likely death if an explosion goes through it.
            for i3:=int8(0);i3<4;i3++ {
              dangerPortals++
              nextPortal:=getPortalSquare(b,p,i2,i3)
              var Portalz interface{}
              Portalz=board.Portal{nextPortal/11,nextPortal%11,(i3+2)%4}
              queue.AddToQueue(&p11,&Portalz)
            }
            //Opponent can shoot a portal in the direction they are facing.
            //They can likely escape if they do so.
            safePortals++
            nextPortal:=getPortalSquare(b,p,int8(i2),(*reachableSquares)[i2].facing)
            var Portalz interface{}
            Portalz=board.Portal{int8(nextPortal/11),int8(nextPortal%11),int8(((*reachableSquares)[i2].facing+2)%4)}
            queue.AddToQueue(&p21,&Portalz)
          }
          if (((*reachableSquares)[i2].dist<i-2)) {
            //Opponent can shoot a portal in any direction.
            //They can likely escape if they do so.
            for i3:=int8(0);i3<4;i3++ {
              safePortals++
              nextPortal:=getPortalSquare(b,p,i2,i3)
              var Portalz interface{}
              Portalz=board.Portal{nextPortal/11,nextPortal%11,(i3+2)%4}
              queue.AddToQueue(&p21,&Portalz)
            }
          }
        }
      }
    }
    //We have the list of all additional possible bombs.
    //First add the "safe" bombs
    tempActiveBombs:=make([]board.Bomb,len(activeBombs)+safeBombs)
    for j:=0;j< len(activeBombs);j++ {
      tempActiveBombs[j]=activeBombs[j];
    }
    for j:=0;j< safeBombs;j++ {
      tempActiveBombs[j+len(activeBombs)]=(*queue.GetFromQueue(&b2)).(board.Bomb);
    }
    activeBombs=tempActiveBombs
    activeBombs=append(activeBombs,p.Bombs...)
    //Now add the safe portals
    currentPortals:=0;
    if (p.O.BluePortal!=nil) {currentPortals++}
    if (p.O.OrangePortal!=nil) {currentPortals++}
    safePortalList:=make([]board.Portal,safePortals+currentPortals)
    for j:=0;j<safePortals;j++ {
      safePortalList[j]=(*queue.GetFromQueue(&p21)).(board.Portal);
    }
    if (p.O.BluePortal!=nil) {safePortalList[len(safePortalList)-1]=(*p.O.BluePortal)}
    if (p.O.OrangePortal!=nil&&p.O.BluePortal==nil) {safePortalList[len(safePortalList)-1]=(*p.O.OrangePortal)}
    if (p.O.OrangePortal!=nil&&p.O.BluePortal!=nil) {safePortalList[len(safePortalList)-2]=(*p.O.OrangePortal)}
    //We must test all combinations of portals
    for k1:=0;k1< len(safePortalList);k1++ {
      for k2:=0;k2< len(safePortalList);k2++ {
        //Two portals can't be the same
        if (k1==k2) {continue}
        pisces.O.BluePortal=&safePortalList[k1]
        pisces.O.OrangePortal=&safePortalList[k2]
        //Portals configured, re-detonate all bombs
        copy(pisces.Bombs,activeBombs)
        for k3:=len(tempActiveBombs)-1;k3>=0 ;k3-- {
          board.ProcessExplosion(b,&pisces,&activeBombs[k3],&d1,&d2)
        }
        //Detonate all bombs touching apocalypse Trails
        if (apo>0) {
          for trailCount:=int8(0);trailCount<121;trailCount++ {
            if (apoBools[trailCount]) {
              tempBomb:=board.BombAt(&pisces,trailCount/11,trailCount%11)
              if tempBomb!=nil {
                board.ProcessExplosion(b,&pisces,tempBomb,&d1,&d2)
              }
            }
          }
        }
        for k3:=0;k3<121;k3++ {
          //Now we mark any blocked squares as marginally unsafe
          if (pisces.Trails[k3]&&(*reachableSquares)[k3].unsafe[i]<2) {(*reachableSquares)[k3].unsafe[i]=2}
        }
      }
    }
    //We now consider suicide attacks.
    //First add the "unsafe" bombs
    tempActiveBombs2:=make([]board.Bomb,len(tempActiveBombs)+dangerBombs)
    for j:=0;j< len(tempActiveBombs);j++ {
      tempActiveBombs2[j]=tempActiveBombs[j];
    }
    for j:=0;j< dangerBombs;j++ {
      tempActiveBombs2[j+len(tempActiveBombs)]=(*queue.GetFromQueue(&b1)).(board.Bomb);
    }
    activeBombs=tempActiveBombs2
    activeBombs=append(activeBombs,p.Bombs...)
    //Now add the unsafe portals
    unsafePortalList:=make([]board.Portal,len(safePortalList)+dangerPortals)
    for j:=0;j< len(safePortalList);j++ {
      unsafePortalList[j]=safePortalList[j]
    }
    for j:=0;j< dangerPortals;j++ {
      unsafePortalList[j+len(safePortalList)]=(*queue.GetFromQueue(&p11)).(board.Portal);
    }
    //We must test all combinations of portals
    for k1:=0;k1< len(unsafePortalList);k1++ {
      for k2:=0;k2< len(unsafePortalList);k2++ {
        //Two portals can't be the same
        if (k1==k2) {continue}
        pisces.O.BluePortal=&unsafePortalList[k1]
        pisces.O.OrangePortal=&unsafePortalList[k2]
        //Portals configured, re-detonate all bombs
        copy(pisces.Bombs,activeBombs)
        for k3:=len(tempActiveBombs)-1;k3>=0;k3-- {
          board.ProcessExplosion(b,&pisces,&activeBombs[k3],&d1,&d2)
        }
        //Detonate all bombs touching apocalypse Trails
        if (apo>0) {
          for trailCount:=int8(0);trailCount<121;trailCount++ {
            if (apoBools[trailCount]) {
              tempBomb:=board.BombAt(&pisces,trailCount/11,trailCount%11)
              if tempBomb!=nil {
                board.ProcessExplosion(b,&pisces,tempBomb,&d1,&d2)
              }
            }
          }
        }
        for k3:=0;k3<121;k3++ {
          //Now we mark any blocked squares as safe barring a suicide attack
          if (pisces.Trails[k3]&&(*reachableSquares)[k3].unsafe[i]<1) {(*reachableSquares)[k3].unsafe[i]=1}
        }
      }
    }
  }
  for i:=5;i>0;i-- {
    for i2:=0;i2<121;i2++ {
      if ((*reachableSquares)[i2].owner==&p.P)&&(((*reachableSquares)[i2].unsafe[i-1])>((*reachableSquares)[i2].unsafe[i])) {
        (*reachableSquares)[i2].unsafe[i]=(*reachableSquares)[i2].unsafe[i-1]
      }
    }
  }
  //Another weird case
  for i2:=0;i2<121;i2++ {
    if p.Trails[i2] {(*reachableSquares)[i2].unsafe[1]=3}
  }
}
func getPaths(b* board.Board,p *board.Pieces,reachableSquares *[]SquareData,paths *[][]int8,maxDanger int,apo int) {
  graph:=SDGraphHolder{reachableSquares,make([][]Edge,len(*reachableSquares))}
  for x:=int8(0);x< int8(len(*reachableSquares));x++ {
    //Edge to self
    cDists:=[]int8{-11,-1,1,11}
    for y:=0;y<4;y++ {
      if ((0>x+cDists[y])||(121<=x+cDists[y])) {continue}
      if ((*reachableSquares)[x].owner)==&p.P {
        if ((*reachableSquares)[x+cDists[y]].owner)==&p.P {
          graph.e[x]=append(graph.e[x],Edge{x+cDists[y],1,nil})
        }
      }
    }
    graph.e[x]=append(graph.e[x],Edge{x,1,&[]int8{0}})
    if ((*reachableSquares)[x].src!=0) {
      src:=(*reachableSquares)[x].src
      weight:=(*reachableSquares)[x].dist-(*reachableSquares)[src].dist
      var moves []int8
      //Check the four basic directional moves
      if (x-src==-1) {moves=[]int8{1};weight=1
      } else if (x-src==-11) {moves=[]int8{4};weight=1
      } else if (x-src==1) {moves=[]int8{3};weight=1
      } else if (x-src==11) {moves=[]int8{2};weight=1
      }
      graph.e[src]=append(graph.e[src],Edge{x,weight,&moves})
      if (weight==1&&(len(moves)>0)&&moves[0]!=0) {
        //Backtracking possible
        moves1:=[]int8{((moves[0]+2)%4)}
        graph.e[x]=append(graph.e[x],Edge{src,1,&moves1})
      }
    }
  }
  start:=p.P.X*11+p.P.Y
  //For each danger level
  for x:=int8(0);x< int8(maxDanger);x++ {
    //A path is a sequence of squares beginning with the start square and ending on the destination square
    //newpaths[length of path][path #][square #]
    newPaths:=make([][][]int8,6)
    newPaths[0]=make([][]int8,1)
    newPaths[0][0]=[]int8{start}
    q:=make([]queue.Queue,6)
    numNewPaths:=make([]int,6)
    //For each path length
    for y:=int8(0);y<5;y++ {
      apoBools:=apocalypseTrails(apo+int(y))
      //For each path
      for z:=0;z< len(newPaths[y]);z++ {
        //Reject this path if we're standing on an apocalypse trail
        if apoBools[newPaths[y][z][len(newPaths[y][z])-1]] {continue}
        //For each edge beginning in last Square of newPaths[y][z]
        for w:=0;w< len(graph.e[newPaths[y][z][len(newPaths[y][z])-1]]);w++ {
          //Check each edge. If the danger is sufficiently low, add the edge to the path.
          nextDest:=graph.e[newPaths[y][z][len(newPaths[y][z])-1]][w].dest
          nextWeight:=graph.e[newPaths[y][z][len(newPaths[y][z])-1]][w].weight
            //Handling weird bugs
          if ((y==0)&&(p.Trails[nextDest])&&(nextWeight==1)) {continue}
          if ((((p.O.X==nextDest/11)&&(p.O.Y==nextDest%11)))) {continue}
          if ((board.BombAt(p,nextDest/11,nextDest%11)!=nil)&&(board.BombAt(p,nextDest/11,nextDest%11).Ticks+2>y)) {continue}
          if ((y+nextWeight<6)&&
          ((*reachableSquares)[nextDest].unsafe[y+nextWeight]<=x)) {
            //Check if weight>1, we must wait on current square for weight-1 turns
            //Make sure this is safe if this is the case

            if nextWeight>1 {
              unsafe:=false
              for v:=int8(1);v<nextWeight;v++ {
                if ((*reachableSquares)[newPaths[y][z][len(newPaths[y][z])-1]].unsafe[y+v]>x) {unsafe=true}
              }
              if (unsafe) {continue};
            }
            numNewPaths[y+nextWeight]++
            nextPath:=make([]int8,len(newPaths[y][z]))
            copy(nextPath,newPaths[y][z])
            nextPath=append(nextPath,nextDest)
            var nextInterface interface{}
            nextInterface=nextPath
            queue.AddToQueue(&q[y+nextWeight],&nextInterface)

          }
        }
      }//End path loop (z)
      newPaths[y+1]=make([][]int8,numNewPaths[y+1])
      for z:=0;z<numNewPaths[y+1];z++ {
        //Add all new paths to the slice
        newPaths[y+1][z]=(*queue.GetFromQueue(&q[y+1])).([]int8)
      }//End of add all new paths loop (z)
    }//End of path length loop (y)
    if len((newPaths[5]))!=0 {
      *paths=make([][]int8,len(newPaths[5]))
      for k:=0;k< len(newPaths[5]);k++ {
        (*paths)[k]=make([]int8,len(newPaths[5][k]))
        copy((*paths)[k],newPaths[5][k])
      }
      return
    }
  }//End danger loop (x)
}
//Return the value of a bomb placed at the end of the path.
//This function does not consider safety.
func pathValue(b *board.Board,p *board.Pieces,path *[]int8,numBlank int,numSoftBlocks int,reachableSquares *[]SquareData,apo int) int {
//First ensure that we do not need to use any soft block portals.
  if (len(*path)<6) {
    //We need to shoot at least one portal. Make sure it's not on a soft block.
    for x:=1;x<len(*path);x++ {
      diff:=((*path)[x])-((*path)[x-1])
      if ((diff!=-11)&&(diff!=-1)&&(diff!=11)&&(diff!=1)) {
        var nearPortal board.Portal
        var farPortal board.Portal
        //We use a portal here
        nearSquare0:=((*path)[x-1])
        farSquare0:=((*path)[x])
        fpColor:=""
        npColor:=""
        //See which portals exist
        directions:=[]int8{-11,-1,11,1}
        for dir:=int8(0);dir<4;dir++ {
          //Check for far portal presence
          farSquare:=farSquare0+directions[dir]
          far:=board.Portal{farSquare/11,farSquare%11,(dir+2)%4}
          if (board.PortalEquals(&far,p.P.BluePortal)) {fpColor="b";farPortal=far}
          if (board.PortalEquals(&far,p.P.OrangePortal)) {fpColor="o";farPortal=far}
          //Check for near portal presence or ability to put a near portal
          nearSquare:=nearSquare0+directions[dir]
          near:=board.Portal{nearSquare/11,nearSquare%11,(dir+2)%4}
          if (board.PortalEquals(&near,p.P.BluePortal)) {npColor="b";nearPortal=near}
          if (board.PortalEquals(&near,p.P.OrangePortal)) {npColor="o";nearPortal=near}
          if (b.HardBlockBoard[nearSquare]&&(npColor=="")) {
            npColor="y"
            nearPortal=near
          }
        }
        if (npColor=="") {
          //We can't shoot a near portal on a hard block.
          return -50
        }
        if (fpColor=="") {
          //Make sure we can shoot a far portal
          var dir int8
          if (farSquare0-nearSquare0>10) {dir=2
          } else if (farSquare0-nearSquare0<int8(-10)) {dir=0
          } else if (farSquare0-nearSquare0>0) {dir=3
          } else {dir=1}
          nextPortal:=getPortalSquare(b,p,nearSquare0,dir)
          if (p.SoftBlockBoard[nextPortal]) {return -50
          } else {farPortal=board.Portal{nextPortal/11,nextPortal%11,(dir+2)%4}}
        }
        //If we made it here we can shoot the portals needed.
        //Ensure that it is safe to do so.
        tempBlue:=p.P.BluePortal
        tempOrange:=p.P.OrangePortal
        p.P.BluePortal=&nearPortal
        p.P.OrangePortal=&farPortal
        rsTemp:=make([]SquareData,121)
        for sc:=0;sc<121;sc++ {
          copySquareData(&rsTemp[sc],(*reachableSquares)[sc])
        }
        mark4Death(b,p,&rsTemp,apo)
        //Check if the path is safe
        waitTime:=6-len(*path)
        for safety:=0;safety<=x;safety++ {
          if (rsTemp[(*path)[safety]].unsafe[safety]>0) {
            //revert state of pieces
            p.P.BluePortal=tempBlue
            p.P.OrangePortal=tempOrange
            return -50
          }
        }
        for wait:=0;wait<waitTime;wait++ {
          if (rsTemp[(*path)[x]].unsafe[x+wait+1]>0) {
            //revert state of pieces
            p.P.BluePortal=tempBlue
            p.P.OrangePortal=tempOrange
            return -50
          }
        }
        for endSeq:=5;(endSeq>x+waitTime);endSeq-- {
          if (rsTemp[(*path)[len((*path))-1-(5-endSeq)]].unsafe[endSeq]>0) {
            p.P.BluePortal=tempBlue
            p.P.OrangePortal=tempOrange
            return -50
          }
        }
        //revert state of pieces
        p.P.BluePortal=tempBlue
        p.P.OrangePortal=tempOrange
      }
    }
  }
  destNum:=len(*path)
  for x:=0;x<numBlank;x++ {
    destNum--
  }
  bomb:=board.Bomb{(*path)[destNum]/11,(*path)[destNum]%11,true,5}
  pisces:=board.Pieces{}
  board.CopyPieces(&pisces,p)
  value:=0
  if (numSoftBlocks>10) {value=1*numBlank //Favor placing the bomb sooner
  } else {value=10*numBlank}
  destroyedByMe:=make([]bool,121)
  destroyedByOpp:=make([]bool,121)
  for len(pisces.Bombs)>0 {
    board.ProcessExplosion(b,&pisces,&pisces.Bombs[0],&destroyedByMe,&destroyedByOpp)
      if (len(p.Bombs)>0) {pisces.Bombs=append(pisces.Bombs[1:])} else {pisces.Bombs=[]board.Bomb{}}
  }
  pisces.Bombs=append(pisces.Bombs,bomb)
  destroyedByMe=make([]bool,121)
  destroyedByOpp=make([]bool,121)
  board.ProcessExplosion(b,&pisces,&pisces.Bombs[0],&destroyedByMe,&destroyedByOpp)
  //Add up value of blocks that would be destroyed by this bomb
  for i:=int8(0);i < int8(len(destroyedByMe));i++ {
    if (destroyedByMe[i]) {
      value+=int(board.BlockValue(b,i/b.BoardSize,i%b.BoardSize))
    }
  }
  dest:=(*path)[len(*path)-1]
  if (numSoftBlocks<6) {
    //If there are few or no soft blocks left, tend towards the center.
    value+=10*int(board.BlockValue(b,dest/11,dest%11))
  }
  return value
}
func GetMove(h *Harambe) string {
  safeForBomb:=false
  b:=&h.CurrentBoard
  p:=&h.CurrentPieces
  if (h.Apocalypse<=0) {
    if (p.Trails[10]) {
      h.Apocalypse=1
      log.Println("The apocalypse is upon us.")
    }
  } else {
    h.Apocalypse++
    log.Println("The apocalypse is upon us.",h.Apocalypse)
  }
  numSoftBlocks:=0
  for sc:=0;sc<121;sc++ {
    if (p.SoftBlockBoard[sc]) {numSoftBlocks++}
  }
  reachableSquares:=make([]SquareData,121)
  getReachableSquares(b,p,5,&reachableSquares)
  mark4Death(b,p,&reachableSquares,h.Apocalypse)
  var paths [][]int8
  getPaths(b,p,&reachableSquares,&paths,3,h.Apocalypse)
  values:=make([]int,len(paths))
  chosenPath:=0
  chosenValue:=0
  for x:=0;x<len(paths);x++ {
    //How many identical numbers are at the end?
    lengthEndSequence:=1
    for y:=len(paths[x])-1;y>0;y-- {
      if paths[x][y]==paths[x][y-1] {lengthEndSequence++} else {break}
    }
    if (lengthEndSequence==6) {safeForBomb=true}
    values[x]=pathValue(b,p,&(paths[x]),lengthEndSequence,numSoftBlocks,&reachableSquares,h.Apocalypse)
  }
  for bigOne:=0;bigOne<100;bigOne++ {
    if (safeForBomb) {
      //If this spot is safe to put a bomb see if it's safe to buy an upgrade
      //Try to buy an upgrade.
      if (p.P.Coins>=5) {
        //We can buy an upgrade. Make sure it's safe.
        if ((manhattanDist(&p.P,&p.O)>7)||(freeSquares(b,p,&p.P)>2)) {
          //Buy count first unless we already did.
          if (!h.BoughtCount) {
            h.BoughtCount=true;
            return "buy_count"
          }
          //Alternate range and pierce upgrades.
          if (p.P.BombRange-2>p.P.Pierce) {
            return "buy_pierce"
          } else {return "buy_range"}
        }
      }
    }
    if (bigOne>10) {log.Println("THE BIG ONE IS",bigOne)}
    chosenPath=0
    chosenValue=0
    for x:=0;x<len(paths);x++ {
      if (values[x]>chosenValue) {
        chosenValue=values[x]
        chosenPath=x
      }
    }
    if (len(paths)==0) {
      //No routes that don't result in our death
      if (h.Apocalypse>0) {
        maxPoints:=int16(0)
        bestMove:=int8(0)
        apoBools:=apocalypseTrails(h.Apocalypse+1)
        for aa:=int8(1);aa<5;aa++ {
          pisces:=board.Pieces{}
          board.CopyPieces(&pisces,p)
          if board.ApplyMove(b,&pisces,aa) {
            if apoBools[pisces.P.X*11+pisces.P.Y] {continue}
            points:=board.BlockValue(b,pisces.P.X,pisces.P.Y)
            if (points>maxPoints) {
              maxPoints=points
              bestMove=aa
            }
          }
        }
        if (bestMove!=0) {
          return h.MoveList[bestMove]
        }
        //Default: We will die, try to take opponent with us
        log.Println("I'M TAKING YOU WITH ME")
        return "b"
      } else {
        //Do nothing.
        log.Println("I accept my fate.")
        return ""
      }
      //printStuff(p,&reachableSquares)
    }
    log.Println(paths[chosenPath])
    difference:=paths[chosenPath][1]-paths[chosenPath][0]
    if difference==1 {
      return "md"
    }
    if difference==11 {
      return "mr"
    }
    if difference==-11 {
      return "ml"
    }
    if difference==-1 {
      return "mu"
    }
    if difference!=0 {
      //We are to travel through a portal.
      var nearPortal board.Portal
      npColor:=""
      fpColor:=""
      //Let's see if we need to place either portal
      directions:=[]int8{-11,-1,11,1}
      for dir:=int8(0);dir<4;dir++ {
        //Check for far portal presence
        farSquare:=paths[chosenPath][1]+directions[dir]
        far:=board.Portal{farSquare/11,farSquare%11,(dir+2)%4}
        if (board.PortalEquals(&far,p.P.BluePortal)) {fpColor="b"}
        if (board.PortalEquals(&far,p.P.OrangePortal)) {fpColor="o"}
        //Check for near portal presence
        nearSquare:=paths[chosenPath][0]+directions[dir]
        near:=board.Portal{nearSquare/11,nearSquare%11,(dir+2)%4}
        if (board.PortalEquals(&near,p.P.BluePortal)) {nearPortal=near;npColor="b"}
        if (board.PortalEquals(&near,p.P.OrangePortal)) {nearPortal=near;npColor="o"}
      }
      if ((npColor!="")&&(fpColor!="")) {
        //Both portals are ready. Enter.
        toEnter:=(nearPortal.Orientation+2)%4
        if (toEnter==0) {toEnter=4}
        return h.MoveList[toEnter]
      } else {
        //One or both portals not ready
        //check if we are facing the right direction to shoot a portal
        if (npColor=="") {
          nearSquare:=paths[chosenPath][0]+directions[p.P.Orientation]
          if (h.CurrentBoard.HardBlockBoard[nearSquare]) {
              //We can shoot the near portal in front of us.
              //Make sure we do not override the far portal
                log.Println("Shooting near portal")
              if (fpColor!="b") {return "bp"} else {return "op"}
            } else {
              //We are not facing an appropriate direction to shoot the portal.
              //If the far portal is ready, pivot.
              if (fpColor!="") {
                nearSquare=paths[chosenPath][0]
                for dir:=0;dir<4;dir++ {
                  if (h.CurrentBoard.HardBlockBoard[nearSquare+directions[dir]]) {
                    dir+=4;
                    if (dir==4) {dir=8}
                      log.Println("Pivot for near portal")
                    return h.MoveList[dir]
                  }
                }
              }
            }
        }
        if (fpColor=="") {
          //If we are here, then either:
          //1) We put down our near portal.
          //2) We are not facing the correct direction to shoot our near portal.
          //Now, pivot if necessary or shoot far portal if it's not.
          var dir int8
          farSquare:=paths[chosenPath][1]
          nearSquare:=paths[chosenPath][0]
          if (farSquare-nearSquare>10) {dir=2
          } else if (farSquare-nearSquare<int8(-10)) {dir=0
          } else if (farSquare-nearSquare>0) {dir=3
          } else {dir=1}
          if (dir==p.P.Orientation) {
            //We are facing correct direction, shoot portal
              log.Println("Shooting far portal")
            if (npColor!="b") {return "bp"} else {return "op"}
          } else {
            //We are not facing correct direction, pivot
            dir+=4;
            if (dir==4) {dir=8}
              log.Println("Pivot for far portal")
            return h.MoveList[dir]
          }
        }
      }

    }
    //If we did not return by now, we are to plant a bomb, or perhaps simply stand still.
    plantBomb:=true
    if (p.P.Count==0) {plantBomb=false}
    if (plantBomb) {
      for x:=1;x<len(paths[chosenPath]);x++ {
        if paths[chosenPath][x]!=paths[chosenPath][x-1] {
          plantBomb=false
          break
        }
      }
    }
    if (plantBomb) {
      //We are to plant a bomb, if it is safe to do so.
      //Start by making sure it's safe to do so.
      pisces:=board.Pieces{}
      board.CopyPieces(&pisces,p)
      board.ApplyMove(b,&pisces,9)
      rs2:=make([]SquareData,121)
      getReachableSquares(b,&pisces,5,&rs2)
      if (!p.MoveIterator) {
        //Rare case where time needed to escape bomb not correctly calc'd
        for bb:=0;bb<len(pisces.Bombs);bb++ {
          pisces.Bombs[bb].Ticks--
        }
      }
      //Putting bombs on the edge when the opponent is NOT there
      //AND there are few to no soft blocks is pointless and dangerous.
      if (numSoftBlocks<6) {
        if ((p.P.X==1)||(p.P.Y==1)||(p.P.X==9)||(p.P.Y==9))&&
          (!((p.O.X==1)||(p.O.Y==1)||(p.O.X==9)||(p.O.Y==9))) {
            log.Println("Too dangerous to plant bomb here 1")
          values[chosenPath]=-1 //Let's not try that one again
          continue
        }
      }
      mark4Death(b,&pisces,&rs2,h.Apocalypse+1)
      var paths2 [][]int8
      getPaths(b,&pisces,&rs2,&paths2,1,h.Apocalypse)
      printStuff(&pisces,&rs2)
      board.PrintBoard(b,&pisces,int((time.Now().UnixNano()/1000000)%1000000000))
      if (len(paths2)==0) {
        log.Println("Too dangerous to plant bomb here 2")
        //This spot is not safe to put a bomb
        values[chosenPath]=-1 //Let's not try that one again
        //Go as far from here as possible
        if (h.Apocalypse<=0) {farthestPath(&paths,&values)}
        continue
      } else {
        //Make sure there's at least one ACCEPTABLE path following the bomb
        safePathExists:=false
        for pathCount:=0;pathCount<len(paths2);pathCount++ {
          lengthEndSequence:=1
          for yourMom:=len(paths2[pathCount])-1;yourMom>0;yourMom-- {
            if paths2[pathCount][yourMom]==paths2[pathCount][yourMom-1] {lengthEndSequence++} else {break}
          }
          if (pathValue(b,&pisces,&(paths2[pathCount]),lengthEndSequence,numSoftBlocks,&rs2,h.Apocalypse+1)>0) {
            safePathExists=true;
            break;
          }
        }
        if (safePathExists) {
          //It's safe, plant a bomb
          return "b"
        } else {
          //No safe path
          log.Println("Too dangerous to plant bomb here 3")
          //This spot is not safe to put a bomb
          values[chosenPath]=-1 //Let's not try that one again
          //Go as far from here as possible
          if (h.Apocalypse<=0) {farthestPath(&paths,&values)}
          printStuff(&pisces,&rs2)
          board.PrintBoard(b,&pisces,int((time.Now().UnixNano()/1000000)%1000000000))
          continue
        }
      }
    } else {
      toShootPortal:=true
      if ((p.P.BluePortal!=nil)||(p.P.OrangePortal!=nil)) {toShootPortal=false}
      //The best path apparently starts by standing completely still.
      //Also we can't afford any upgrades.
      //Shoot a portal if we don't have one out, preferably at a hard block.
      if (toShootPortal) {
        //Early game, definitely face a hard block.
        nextPortal:=getPortalSquare(b,p,p.P.X*11+p.P.Y,p.P.Orientation)
        if (!b.HardBlockBoard[nextPortal]) {
          //We are not facing a hard block. Remedy that problem.
          for dir:=int8(0);dir<4;dir++ {
          nextPortal=getPortalSquare(b,p,p.P.X*11+p.P.Y,dir)
            if (b.HardBlockBoard[nextPortal]) {
              //We found a direction to turn
              if (dir==0) {return h.MoveList[8]}
              return h.MoveList[dir+4];
            }
          }
          //No suitable direction, do not shoot portal
          toShootPortal=false
        }
      }
      if (toShootPortal) {
        return "bp"
      }
      //Shooting a portal may be inadvisable.
      //Return a random direction change just to mess with opponent.
      random:=rand.New(rand.NewSource(time.Now().UnixNano()))
      dir:=int8(random.Intn(3))
      if (p.P.Orientation==dir) {dir+=3}
      dir=(dir%4)+4;
      if (dir==4) {dir=8}
      log.Println("Random turn move")
      return h.MoveList[dir]
    }
    //We tried to plant a bomb but it is unsafe to do so.

    //If we made it here without returning, something went wrong.
    log.Println("HELP! I don't know what to do.")
    return ""
  }
  log.Println("HELP! I REALLY don't know what to do.")
  //If we made it here without returning, something REALLY went wrong.
  return ""
}
func printStuff(p *board.Pieces,reachableSquares *[]SquareData)  {
  f, e := os.Create(("boards/Conquest"+strconv.Itoa(int((time.Now().UnixNano()/1000000)%1000000000))+".txt"))
      if e != nil {
          panic(e)
      }
  defer f.Close()
  w := bufio.NewWriter(f)
	for i1:=0;i1<11;i1++ {
		for i2:=0;i2<11;i2++ {
			if ((*reachableSquares)[i1+i2*11].owner==&p.P) {
	      _,err:=w.WriteString("P"+strconv.Itoa(int((*reachableSquares)[i1+i2*11].dist))+" ")
  	    for k:=0;k<6;k++ {_,err=w.WriteString(strconv.Itoa(int((*reachableSquares)[i1+i2*11].unsafe[k])))}
  	      _,err=w.WriteString(" ")
	      if (err!=nil) {panic(err)}
	      w.Flush()
			} else if ((*reachableSquares)[i1+i2*11].owner==&p.O) {
        _,err:=w.WriteString("O"+strconv.Itoa(int((*reachableSquares)[i1+i2*11].dist))+"        ")
        if (err!=nil) {panic(err)}
        w.Flush()
			} else {
        _,err:=w.WriteString("********* ")
        if (err!=nil) {panic(err)}
        w.Flush()
			}
		}
    _,err:=w.WriteString("\n")
    if (err!=nil) {panic(err)}
		w.Flush()
	}
}
