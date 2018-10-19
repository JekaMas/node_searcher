package searcher

import (
	"bytes"
	"sync"

	"github.com/davecgh/go-spew/spew"
)

var NodesByClient = make(map[string][]*Node) // map[client]Node
var NodesByClientLock sync.RWMutex

var NodesMap = make(map[string]*Node) // map[ID]Node
var NodesMapLock sync.RWMutex

func PrintNodesMap() []byte {
	s := bytes.Buffer{}

	NodesByClientLock.RLock()

	for client, nodes := range NodesByClient {
		s.WriteString("Client: " + client + "\n")
		for _, node := range nodes {
			s.WriteString(`"` + node.Enode + "\"\n")
		}
	}
	s.WriteString("\n\n\n\nRaw data:\n\n" + spew.Sdump(NodesByClient))

	NodesByClientLock.RUnlock()

	return s.Bytes()
}
