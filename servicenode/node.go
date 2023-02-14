package servicenode

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	Id int

	LocalData   []int
	UpdatedData []int
	DataSize    int // length of the array
	Lock        *sync.Mutex

	MaxNode      int
	MaxItemValue int // max value of an item in array

	ConsecutiveSuccesses int
	SampleSize           int
	DecisionThreshold    int
	Decided              bool
}

type UpdatedDataResponse struct {
	Data []int
}

func InitAnArrayFrom0ToN(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i
	}
	return result
}

func ShuffleFistKNumbers(arr []int, k int) []int {
	arrLen := len(arr)
	for i := 0; i < k; i++ {
		r := rand.Intn(arrLen)
		arr[r], arr[arrLen-1] = arr[arrLen-1], arr[r]
		arrLen -= 1
		if arrLen == 0 {
			break
		}
	}
	return arr[arrLen:]
}

func GenerateRandomKodeIDs(k int, max int) []int {
	nums := InitAnArrayFrom0ToN(max)
	return ShuffleFistKNumbers(nums, k)
}

func (node *Node) queryANeighbour(nodeID int) []int {
	// nodeID = 0
	// fmt.Println("Query NodeID ", nodeID)
	port := 9000 + nodeID
	client := http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	resp, err := client.Get("http://localhost:" + strconv.Itoa(port) + "/localdata")

	if err == nil {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var data UpdatedDataResponse
			err := json.Unmarshal(body, &data)
			if err == nil {
				return data.Data
			} else {
				fmt.Println(err)
				return nil
			}
		} else {
			fmt.Println(err)
			return nil
		}
	} else {
		fmt.Println(err)
		return nil
	}
}

func (node *Node) queryNeighbours() [][]int {
	nodeIDs := GenerateRandomKodeIDs(node.SampleSize, node.MaxNode)
	sampleData := make([][]int, node.SampleSize)
	var wg sync.WaitGroup
	for i := 0; i < len(nodeIDs); i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			data := node.queryANeighbour(nodeIDs[i])
			sampleData[i] = data
		}()
	}
	wg.Wait()
	return sampleData
}

func (node *Node) combineData(sampleData [][]int) []int {
	transformedData := make([][]int, node.DataSize)
	for i := 0; i < node.DataSize; i++ {
		transformedData[i] = make([]int, node.MaxItemValue)
	}

	// calculate transformedData array
	for i := 0; i < node.DataSize; i++ {
		for j := 0; j < node.SampleSize; j++ {
			transformedData[i][sampleData[j][i]]++
		}
	}

	// calculate updated data
	updatedData := make([]int, node.DataSize)
	for i := 0; i < node.DataSize; i++ {
		max := rand.Intn(node.MaxItemValue)
		for j := 0; j < node.MaxItemValue; j++ {
			if transformedData[i][max] < transformedData[i][j] {
				max = j
			}
		}
		updatedData[i] = max
	}
	return updatedData
}

func (node *Node) checkSameResult(new []int) bool {
	fmt.Println("Transform data", node.UpdatedData, new)
	for i := 0; i < len(new); i++ {
		if node.UpdatedData[i] != new[i] {
			return false
		}
	}
	return true
}

// Consensus processing
func (node *Node) processQueryResult(sampleData [][]int) {
	node.Lock.Lock()
	defer node.Lock.Unlock()
	// fmt.Println("Node", node.Id, " update ", sampleData)
	updatedData := node.combineData(sampleData)
	if node.checkSameResult(updatedData) {
		node.ConsecutiveSuccesses++
		if node.ConsecutiveSuccesses >= node.DecisionThreshold {
			node.Decided = true
		}
	} else {
		node.ConsecutiveSuccesses = 0
	}
	node.UpdatedData = updatedData
}

func (node *Node) randomLocalData() {
	node.LocalData = make([]int, node.DataSize)
	for i := 0; i < len(node.LocalData); i++ {
		node.LocalData[i] = rand.Intn(node.MaxItemValue)
	}
}

func (node *Node) init() {
	rand.Seed(time.Now().UnixNano())
	node.Lock = &sync.Mutex{}
	node.randomLocalData()
	node.UpdatedData = node.LocalData
	fmt.Println("Init data:", node.LocalData)
}

func (node *Node) GetUpdatedData() []int {
	node.Lock.Lock()
	defer node.Lock.Unlock()
	result := node.UpdatedData
	return result
}

func (node *Node) Sync() {
	time.Sleep(time.Second * 5)
	fmt.Println("Start syncing ...", node.Id)
	for !node.Decided {
		sampleData := node.queryNeighbours()
		node.processQueryResult(sampleData)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(20)+10))
	}
	fmt.Println("End syncing ...", node.Id)
	fmt.Println("Final result", node.UpdatedData)
}

func MakeNode(id int, maxNode int, maxItemValue int, dataSize int, sampleSize int, decisionThreshold int) *Node {
	node := &Node{Id: id, MaxNode: maxNode, MaxItemValue: maxItemValue, DataSize: dataSize, SampleSize: sampleSize, DecisionThreshold: decisionThreshold}
	node.init()
	return node
}
