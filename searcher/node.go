package searcher

import (
	"github.com/json-iterator/go"
	"strconv"
)

type Node struct {
	ID   string
	Host string
	Port int

	Client        string
	ClientVersion string
	ClientId      string

	Enode        string
	Capabilities string
}

func (n *Node) Decode(input []byte) {
	var json = jsoniter.ConfigFastest
	json.Unmarshal(input, n)

	n.Enode = "enode://" + n.ID + "@" + n.Host + ":" + strconv.Itoa(n.Port)
}
