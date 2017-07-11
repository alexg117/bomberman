package message


// Message defines the structure of responses from the game server
type Message struct {
	HardBlockBoard []int       `json:"hardBlockBoard"`
	BoardSize      int         `json:"boardSize"`
	GameID         string      `json:"gameID"`
	BombMap        interface{} `json:"bombMap"`
	MoveIterator   int         `json:"moveIterator"`
	PlayerID       string      `json:"playerID"`
	PortalMap      interface{} `json:"portalMap"`
	PlayerIndex    int         `json:"playerIndex"`
	TrailMap       interface{} `json:"trailMap"`
	Player         interface{} `json:"player"`
	State          string      `json:"state"`
	SoftBlockBoard []int       `json:"softBlockBoard"`
	MoveOrder      []int       `json:"moveOrder"`
	Opponent       interface{} `json:"opponent"`
}
