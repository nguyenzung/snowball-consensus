package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nguyenzung/snowball-consensus/servicenode"
)

func main() {

	nodeID, _ := strconv.Atoi(os.Args[1])

	ListSize := 16
	NumOfSampleNode := 18
	DecisionThreshold := 5

	node := servicenode.MakeNode(nodeID, 200, 10, ListSize, NumOfSampleNode, DecisionThreshold)

	http.HandleFunc("/localdata", func(w http.ResponseWriter, r *http.Request) {
		data := node.GetUpdatedData()
		res := &servicenode.UpdatedDataResponse{Data: data}
		raw, err := json.Marshal(res)
		if err == nil {
			w.Write(raw)
		} else {
			w.WriteHeader(500)
		}
	})

	go node.Sync()
	port := strconv.Itoa(9000 + nodeID)
	go http.ListenAndServe("localhost:"+port, nil)

	for {
		time.Sleep(time.Second)
	}

}
