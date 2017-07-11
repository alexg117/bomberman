package queue

type InterfaceHolder struct {
  this        *interface{}
  next        *InterfaceHolder
}
type Queue struct {
  first       *InterfaceHolder
  last        *InterfaceHolder
}
//Queue helper functions
func AddToQueue(q *Queue,c *interface{}) {
  ch:=InterfaceHolder{c,nil}
  if q.last!=nil {q.last.next=&ch}
  q.last=&ch
  if q.first==nil {q.first=&ch}
}
func GetFromQueue(q *Queue) *interface{} {
  if (q.first!=nil) {
    yourInterface:=q.first.this
    q.first=q.first.next
    if q.first==nil {q.last=nil}
    return yourInterface
  }
  return nil
}
