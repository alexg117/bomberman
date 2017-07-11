package board

import (
  "bufio"
  "math/rand"
  "math"
  "os"
  "strconv"
  //"time"
)

type Bomb struct {
  X            int8
  Y            int8
  PlayerOwns   bool //True if we own it
  Ticks        int8
}
type Portal struct {
  X            int8
  Y            int8
  Orientation  int8
}
type Avatar struct {
  X            int8
  Y            int8
  Orientation  int8
  BluePortal   *Portal
  OrangePortal *Portal
  Count        int8
  Pierce       int8
  BombRange    int8
  Coins        int16
}
//Holds permanent aspects of board
type Board struct {
  BoardSize      int8
	HardBlockBoard []bool
  //False if we are referred to as "false" in the move list, true otherwise
  PlayerIndex    bool
}
type Pieces struct { //Holds anything that is non-permanent
  P              Avatar    //Player
  O              Avatar   //Opponent
  Bombs          []Bomb
  Trails         []bool
  MoveOrder      []bool
  MoveIterator   bool
	SoftBlockBoard []bool
}
func DefaultBoard(b *Board) {
  b.BoardSize=11
  b.PlayerIndex=false
  b.HardBlockBoard=make([]bool,b.BoardSize*b.BoardSize)
  i:=int8(0)
  for i<b.BoardSize*b.BoardSize {
    if (i%b.BoardSize==0)||
       (i%b.BoardSize==b.BoardSize-1)||
       (i/b.BoardSize==0)||
       (i/b.BoardSize==b.BoardSize-1)||
       ((i/b.BoardSize)%2==0&&(i%b.BoardSize)%2==0) {
       b.HardBlockBoard[i]=true
     }
    i++
  }
}
func DefaultPieces(b *Board,p *Pieces) {
  p.P=Avatar{1,1,0,nil,nil,1,0,3,0}
  p.O=Avatar{9,9,0,nil,nil,1,0,3,0}
  p.Bombs=nil
  p.Trails=make([]bool,b.BoardSize*b.BoardSize)
  p.MoveOrder=make([]bool,2)
  p.MoveOrder[0]=false;p.MoveOrder[1]=true
  p.MoveIterator=false
  p.SoftBlockBoard=make([]bool,b.BoardSize*b.BoardSize)
  i:=int8(0)
  r:=rand.New(rand.NewSource(3000))
  for i<b.BoardSize*b.BoardSize {
    if ((!b.HardBlockBoard[i])&&r.Float32()<.7) {
      p.SoftBlockBoard[i]=true
    }
    i++
  }
  p.SoftBlockBoard[12]=false
  p.SoftBlockBoard[13]=false
  p.SoftBlockBoard[23]=false
  p.SoftBlockBoard[97]=false
  p.SoftBlockBoard[107]=false
  p.SoftBlockBoard[108]=false
}
func CopyAvatar(dest *Avatar,source *Avatar) {
    dest.Coins=source.Coins;
    dest.BombRange=source.BombRange;
    dest.Pierce=source.Pierce;
    dest.Count=source.Count;
    dest.X=source.X
    dest.Y=source.Y
    dest.OrangePortal=source.OrangePortal
    dest.BluePortal=source.BluePortal
    dest.Orientation=source.Orientation
}
func CopyPieces(dest *Pieces,src *Pieces) {
  CopyAvatar(&(dest.P),&(src.P))
  CopyAvatar(&(dest.O),&(src.O))
  dest.Bombs=make([]Bomb,0)
  i:=0
  for i< len(src.Bombs) {
    dest.Bombs=append(dest.Bombs,Bomb{
      src.Bombs[i].X,src.Bombs[i].Y,src.Bombs[i].PlayerOwns,src.Bombs[i].Ticks})
    i++
  }
  dest.Trails=make([]bool,len(src.Trails))
  copy(dest.Trails,src.Trails)
  dest.MoveOrder=make([]bool,len(src.MoveOrder))
  copy(dest.MoveOrder,src.MoveOrder)
  dest.MoveIterator=src.MoveIterator
  dest.SoftBlockBoard=make([]bool,len(src.SoftBlockBoard))
  copy(dest.SoftBlockBoard,src.SoftBlockBoard)
}
//Returns the value of the block at (x,y)
func BlockValue(b *Board,xi int8,yi int8) int16 {
  x:=float64(xi)
  y:=float64(yi)
  return int16(math.Floor((float64(b.BoardSize) - 1 - x) * x * (float64(b.BoardSize) - 1 - y) * y * 10 / (math.Pow((float64(b.BoardSize) - 1),4) / 16)))
}
//Return true if any bomb is at given coordinates
func BombAt(p *Pieces,x int8,y int8) *Bomb {
  i:=0
  for i< len(p.Bombs) {
    if p.Bombs[i].X==x&&p.Bombs[i].Y==y {
      return &p.Bombs[i]
    }
    i++
  }
  return nil
}
//return true if they have same coordinates and orientation, regardless of owner or color
func PortalEquals(a *Portal,b *Portal) bool {
  if a==nil||b==nil {return false}
  return (a.X==b.X&&a.Y==b.Y&&a.Orientation==b.Orientation)
}
//If there's a portal with orientaiton, return the exit portal if any.
//If no portal or the portal there has wrong orientation or no exit portal return nil.
func PlayerAt(p *Pieces,x int8,y int8) bool {
  if ((p.P.X==x&&p.P.Y==y))||
     ((p.O.X==x&&p.O.Y==y)) {
     return true
   }
   return false
}
func GetPortal(p *Pieces,x int8,y int8,orientation int8) *Portal {
  temp:=Portal{x,y,orientation}
  if PortalEquals(p.P.BluePortal,&temp) {return p.P.OrangePortal}
  if PortalEquals(p.P.OrangePortal,&temp) {return p.P.BluePortal}
  if PortalEquals(p.O.BluePortal,&temp) {return p.O.OrangePortal}
  if PortalEquals(p.O.OrangePortal,&temp) {return p.O.BluePortal}
  return nil
}
//If these coordinates and orientation match a portal, return said portal.
func HasPortal(p *Pieces,x int8,y int8,orientation int8) *Portal {
  temp:=Portal{x,y,orientation}
  if PortalEquals(p.P.BluePortal,&temp) {return p.P.BluePortal}
  if PortalEquals(p.P.OrangePortal,&temp) {return p.P.OrangePortal}
  if PortalEquals(p.O.BluePortal,&temp) {return p.O.BluePortal}
  if PortalEquals(p.O.OrangePortal,&temp) {return p.O.OrangePortal}
  return nil
}
func ProcessExplosion(b *Board,p *Pieces,bomb *Bomb,destroyedByPlayer *[]bool,destroyedByOpponent *[]bool) {
  isPlayer:=bomb.PlayerOwns
  var active *Avatar
  if (isPlayer) {active=&(p.P)
  } else {active=&(p.O)}
  p.Trails[b.BoardSize*bomb.X+bomb.Y]=true
  if (bomb.PlayerOwns) {p.P.Count++
    } else {p.O.Count++}
  i:=int8(0)
  for i<4 {
    dir:=i
    x:=bomb.X
    y:=bomb.Y
    r:=active.BombRange
    pierce:=int8(100) //Don't count down pierce for real till we hit a block
    for (r>0)&&(pierce>0)&&(0<x)&&(x<b.BoardSize-1&&0<y)&&(y<b.BoardSize-1) {
      //Move to next block
      if dir==0 {x--}
      if dir==1 {y--}
      if dir==2 {x++}
      if dir==3 {y++}
      //If we hit a portal and there's another it leads to, go there
      portalTemp:=GetPortal(p,x,y,(dir+2)%4)
      if (portalTemp!=nil) {
        x=portalTemp.X
        y=portalTemp.Y
        dir=portalTemp.Orientation
        if dir==0 {x--}
        if dir==1 {y--}
        if dir==2 {x++}
        if dir==3 {y++}
      }
      //If we hit a bomb, detonate it
      chainBomb:=BombAt(p,x,y)
      if (chainBomb!=nil&&!p.Trails[b.BoardSize*x+y]) {
        ProcessExplosion(b,p,chainBomb,destroyedByPlayer,destroyedByOpponent)
        //Since this bomb is exploding, remove it from the Pieces object
        i2:=0
        for i2< len(p.Bombs) {
          if (p.Bombs[i2].X==x&&p.Bombs[i2].Y==y) {
            p.Bombs=append(p.Bombs[:i2], p.Bombs[i2+1:]...)
            break
          }
          i2++
        }
      }
      //Mark this spot as part of a trail
      p.Trails[b.BoardSize*x+y]=true
      //If we hit a soft block, record that we destroyed it
      if (p.SoftBlockBoard[x*b.BoardSize+y]) {
        if (isPlayer) {
          (*destroyedByPlayer)[x*b.BoardSize+y]=true
        } else {
          (*destroyedByOpponent)[x*b.BoardSize+y]=true
        }
      }
      //Drop range and pierce by 1 since we advanced
      r--
      pierce--
      //If we hit a block or a player, start counting down pierce for real, if we aren't already
      if (b.HardBlockBoard[b.BoardSize*x+y]||
        p.SoftBlockBoard[b.BoardSize*x+y]||
        PlayerAt(p,x,y))&&
        pierce>50 {
        pierce=active.Pierce
      }
    }
    i++
  }
}
//Advance iterator and check for new explosions
func IterateMove(b *Board,p *Pieces,moveCode int8) {
  destroyedByPlayer:=make([]bool,b.BoardSize*b.BoardSize)
  destroyedByOpponent:=make([]bool,b.BoardSize*b.BoardSize)
  i:=int8(0)
  if (!p.MoveIterator) {
    p.MoveIterator=true
  } else {
    p.MoveIterator=false;
    temp:=p.MoveOrder[0]
    p.MoveOrder[0]=p.MoveOrder[1]
    p.MoveOrder[1]=temp
    p.Trails=make([]bool,b.BoardSize*b.BoardSize)
    for i< int8(len(p.Bombs)) {
      p.Bombs[i].Ticks--
      if p.Bombs[i].Ticks==0 {
        ProcessExplosion(b,p,&p.Bombs[i],&destroyedByPlayer,&destroyedByOpponent)
        p.Bombs=append(p.Bombs[:i], p.Bombs[i+1:]...)
      } else {i++}
    }
  }
  i=0
  //Process block destruction and award coins
  for i<int8(len(p.SoftBlockBoard)) {
    if (destroyedByPlayer[i]) {
      p.SoftBlockBoard[i]=false
      p.P.Coins+=BlockValue(b,i/b.BoardSize,i%b.BoardSize)
    }
    if (destroyedByOpponent[i]) {
      p.SoftBlockBoard[i]=false
      p.O.Coins+=BlockValue(b,i/b.BoardSize,i%b.BoardSize)
    }
    //Remove portals whose blocks are destroyed
    if (p.P.BluePortal!=nil)&&(!b.HardBlockBoard[p.P.BluePortal.X*11+p.P.BluePortal.Y])&&
          !(p.SoftBlockBoard[p.P.BluePortal.X*11+p.P.BluePortal.Y]) {p.P.BluePortal=nil}
    if (p.P.OrangePortal!=nil)&&(!b.HardBlockBoard[p.P.OrangePortal.X*11+p.P.OrangePortal.Y])&&
          !(p.SoftBlockBoard[p.P.OrangePortal.X*11+p.P.OrangePortal.Y]) {p.P.OrangePortal=nil}
    if (p.O.BluePortal!=nil)&&(!b.HardBlockBoard[p.O.BluePortal.X*11+p.O.BluePortal.Y])&&
          !(p.SoftBlockBoard[p.O.BluePortal.X*11+p.O.BluePortal.Y]) {p.O.BluePortal=nil}
    if (p.O.OrangePortal!=nil)&&(!b.HardBlockBoard[p.O.OrangePortal.X*11+p.O.OrangePortal.Y])&&
          !(p.SoftBlockBoard[p.O.OrangePortal.X*11+p.O.OrangePortal.Y]) {p.O.OrangePortal=nil}
    i++
  }
}
//Return true if this move is not effectively a "pass"
//If this function would return true, call the IterateMove function.
func ApplyMove(b *Board,p *Pieces,moveCode int8) bool {
  moved:=false
  var isPlayer bool
  if (p.MoveIterator) {isPlayer=p.MoveOrder[1]==b.PlayerIndex} else
  {isPlayer=p.MoveOrder[0]==b.PlayerIndex}
  var active *Avatar
  if (isPlayer) {active=&(p.P)
  } else {active=&(p.O)}
  switch moveCode {

  case 1,2,3,4: //Move character, return true if move not blocked
    x:=active.X
    y:=active.Y
    newOrientation:=moveCode%4
    switch moveCode {
    case 1:
      y--
    case 2:
      x++
    case 3:
      y++
    case 4:
      x--
    }
    destPortal:=GetPortal(p,x,y,(moveCode+2)%4)
    if (destPortal!=nil) {
      x=destPortal.X
      y=destPortal.Y
      newOrientation=destPortal.Orientation
      switch destPortal.Orientation {
      case 0:
        x--
      case 1:
        y--
      case 2:
        x++
      case 3:
        y++
      }
    }
    if (b.HardBlockBoard[b.BoardSize*x+y])||
    (p.SoftBlockBoard[b.BoardSize*x+y])||
    (p.Trails[b.BoardSize*x+y])||
    (BombAt(p,x,y)!=nil)||
    (PlayerAt(p,x,y)) { //Blocked by block, bomb or explosion
      return false
    } else {
      active.X=x
      active.Y=y
      active.Orientation=newOrientation
      moved= true
    }
  case 5,6,7,8:
    newOrientation:=moveCode%4
    if (isPlayer) {
      if newOrientation==p.P.Orientation {return false
      } else {p.P.Orientation=newOrientation}
    } else if (newOrientation==p.O.Orientation) {return false
    } else {p.O.Orientation=newOrientation }
    moved= true
  case 9://Bomb
    if isPlayer&&(BombAt(p,p.P.X,p.P.Y)!=nil||p.P.Count==0) {return false
    } else if (!isPlayer)&&(BombAt(p,p.P.X,p.P.Y)!=nil||p.O.Count==0) {return false}
    if (isPlayer) {
      p.Bombs=append(p.Bombs,Bomb{p.P.X,p.P.Y,true,5})
    } else {
      p.Bombs=append(p.Bombs,Bomb{p.O.X,p.O.Y,false,5})
    }
    active.Count--
    moved= true
  case 10,11://Portal
    x:=active.X
    y:=active.Y
    orientation:=active.Orientation
    //Find the block that this portal will stick to

    for (!b.HardBlockBoard[b.BoardSize*x+y])&&!p.SoftBlockBoard[b.BoardSize*x+y] {
      switch orientation {
      case 0:
        x--
      case 1:
        y--
      case 2:
        x++
      case 3:
        y++
      }
    }
    //Check to see if we're just replacing the same portal
    ptemp:=HasPortal(p,x,y,(orientation+2)%4)
    if (ptemp!=nil) {
      if (moveCode==10) { //Blue Portal
        if (active.BluePortal==ptemp) {
          return false
        }
      } else { //Orange Portal
        if (active.OrangePortal==ptemp) {
          return false
        }
      }
    }
    //If we hit another portal, replace it
    if (p.P.BluePortal!=nil)&&((p.P.BluePortal.X==x)&&(p.P.BluePortal.Y==y)) {p.P.BluePortal=nil}
    if (p.P.OrangePortal!=nil)&&((p.P.OrangePortal.X==x)&&(p.P.OrangePortal.Y==y)) {p.P.OrangePortal=nil}
    if (p.O.BluePortal!=nil)&&((p.O.BluePortal.X==x)&&(p.O.BluePortal.Y==y)) {p.O.BluePortal=nil}
    if (p.O.OrangePortal!=nil)&&((p.O.OrangePortal.X==x)&&(p.O.OrangePortal.Y==y)) {p.O.OrangePortal=nil}
    //If we made it this far, shooting a portal will not hit our portal of the same color
    if (moveCode==10) {
      active.BluePortal=&Portal{x,y,(orientation+2)%4}
    } else {
      active.OrangePortal=&Portal{x,y,(orientation+2)%4}
    }
    moved= true
  case 12,13,14: //Buy upgrade, but not a block, that's 15
    if active.Coins==0 {return false
    } else {
      active.Coins--
      switch moveCode {
      case 12:
        active.Count++
      case 13:
        active.Pierce++
      case 14:
        active.BombRange++
      }
    }
    moved=true
  case 15:
    x:=active.X
    y:=active.Y
    switch active.Orientation {
    case 1:
      y--
    case 2:
      x++
    case 3:
      y++
    case 4:
      x--
    }
    //Make sure there isn't already a block, bomb, or player there
    if (b.HardBlockBoard[b.BoardSize*x+y])||
    (p.SoftBlockBoard[b.BoardSize*x+y])||
    (BombAt(p,x,y)!=nil)||
    (PlayerAt(p,x,y)) {return false} //Blocked by block, bomb or explosion
    //Make sure we have enough money
    bv:=BlockValue(b,x,y)
    if active.Coins<bv {return false}
    //Everything looks good, buy the block
    active.Coins-=bv
    p.SoftBlockBoard[x*b.BoardSize+y]=true
    moved=true
  }
  if moved {IterateMove(b,p,moveCode)}
  return moved;
}
func PrintBoard(b *Board,p *Pieces,count int) {
  f, e := os.Create(("boards/board"+strconv.Itoa(count)+".txt"))
      if e != nil {
          panic(e)
      }
  defer f.Close()
  w := bufio.NewWriter(f)
  //Row iterator
  i1:=int8(0)
  for i1<b.BoardSize {
    //4 sub-rows per row rep
    i2:=int8(0)
    for i2<4 {
      line:=""
      //Column iterator
      i3:=int8(0)
      for i3<b.BoardSize {
        if (i2==0) {line+="+---"
        } else {
          line+="|"
          if (i2==1) {
            portal:=HasPortal(p,i3,i1,1)
            player:=PlayerAt(p,i3,i1)
            if (portal!=nil) {
              if (PortalEquals(portal,p.P.OrangePortal)) {line+=" O "}
              if (PortalEquals(portal,p.O.OrangePortal)) {line+=" o "}
              if (PortalEquals(portal,p.P.BluePortal)) {line+=" B "}
              if (PortalEquals(portal,p.O.BluePortal)) {line+=" b "}
            } else if (player) {
              if ((p.P.X==i3)&&(p.P.Y==i1)&&(p.P.Orientation==1)) {line+=" F "} else
              if ((p.O.X==i3)&&(p.O.Y==i1)&&(p.O.Orientation==1)) {line+=" F "} else
              {line+="   "}
            } else {line+="   "}
          } else if (i2==2) {
            portal1:=HasPortal(p,i3,i1,0)
            portal2:=HasPortal(p,i3,i1,2)
            player:=PlayerAt(p,i3,i1)
            if (portal1!=nil) {
              if (PortalEquals(portal1,p.P.OrangePortal)) {line+="O"}
              if (PortalEquals(portal1,p.O.OrangePortal)) {line+="o"}
              if (PortalEquals(portal1,p.P.BluePortal)) {line+="B"}
              if (PortalEquals(portal1,p.O.BluePortal)) {line+="b"}
            } else if (player) {
              if ((p.P.X==i3)&&(p.P.Y==i1)&&(p.P.Orientation==0)) {line+="F"} else
              if ((p.O.X==i3)&&(p.O.Y==i1)&&(p.O.Orientation==0)) {line+="F"} else
              {line+=" "}
            } else {line+=" "}
            if (b.HardBlockBoard[i3*b.BoardSize+i1]) {line+="H"} else
            if (p.SoftBlockBoard[i3*b.BoardSize+i1]) {line+="S"} else
            if (p.Trails[i3*b.BoardSize+i1]) {line+="T"} else
            if (BombAt(p,i3,i1)!=nil) {line+=strconv.Itoa(int(BombAt(p,i3,i1).Ticks))} else
            if (player) {
              if ((p.P.X==i3)&&(p.P.Y==i1)) {line+="@"} else
              if ((p.O.X==i3)&&(p.O.Y==i1)) {line+="#"}
            } else  {
              line+=" "
            }
            if (portal2!=nil) {
              if (PortalEquals(portal2,p.P.OrangePortal)) {line+="O"}
              if (PortalEquals(portal2,p.O.OrangePortal)) {line+="o"}
              if (PortalEquals(portal2,p.P.BluePortal)) {line+="B"}
              if (PortalEquals(portal2,p.O.BluePortal)) {line+="b"}
            } else if (player) {
              if ((p.P.X==i3)&&(p.P.Y==i1)&&(p.P.Orientation==2)) {line+="F"} else
              if ((p.O.X==i3)&&(p.O.Y==i1)&&(p.O.Orientation==2)) {line+="F"} else
              {line+=" "}
            } else {line+=" "}
          } else { //i2==3
            portal:=HasPortal(p,i3,i1,3)
            player:=PlayerAt(p,i3,i1)
            if (portal!=nil) {
              if (PortalEquals(portal,p.P.OrangePortal)) {line+=" O "}
              if (PortalEquals(portal,p.O.OrangePortal)) {line+=" o "}
              if (PortalEquals(portal,p.P.BluePortal)) {line+=" B "}
              if (PortalEquals(portal,p.O.BluePortal)) {line+=" b "}
            } else if (player) {
              if ((p.P.X==i3)&&(p.P.Y==i1)&&(p.P.Orientation==3)) {line+=" F "} else
              if ((p.O.X==i3)&&(p.O.Y==i1)&&(p.O.Orientation==3)) {line+=" F "} else
              {line+="   "}
            } else {line+="   "}
          }
        }
        i3++
      }
      if i2==0 {line+="+\n"} else {line+="|\n"}
      _,err:=w.WriteString(line)
      if (err!=nil) {panic(err)}
      w.Flush()
      i2++
    }
    i1++
  }
  _,err:=w.WriteString("P1")
  _,err=w.WriteString("Coins "+strconv.Itoa(int(p.P.Coins)))
  _,err=w.WriteString("P2")
  _,err=w.WriteString("Coins "+strconv.Itoa(int(p.O.Coins)))
  if (err!=nil) {panic(err)}
  w.Flush()
}
