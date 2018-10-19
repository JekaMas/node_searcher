package searcher

import (
	"github.com/json-iterator/go"
	"strconv"
)

type ReqNodes struct {
	Data []*Node
}

func (r *ReqNodes) Decode(input []byte) {
	var json = jsoniter.ConfigFastest
	json.Unmarshal(input, r)

	for i := range r.Data {
		n := r.Data[i]
		n.Enode = "enode://" + n.ID + "@" + n.Host + ":" + strconv.Itoa(n.Port)
	}
}
