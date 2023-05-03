package main

import (
	"fmt"
	"net/http"
	"io"
	"strconv"
	"encoding/json"
	"sort"
)

type Transaction struct {
	Txid string	`json:"txid"`
	Vin []InputTransaction `json:"vin`
}

type InputTransaction struct {
	Txid string	`json:"txid`
}

func main() {
	var blockId int = 680000
	blockHash, _ := HttpGetData("https://blockstream.info/api/block-height/" + strconv.Itoa(blockId))

	transactionMap := BuildTransactionMap(blockHash)

	ancestorMap := BuildAncestorTree(transactionMap)

	SortByValAndPrint(ancestorMap, 10)
}

func HttpGetData(url string) (string, int){
	resp, err := http.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return string(body), resp.StatusCode
}

func BuildTransactionMap(blockHash string) map[string][]string {
	transactionMap := make(map[string][]string)
	var code int
	var response string
	url := "https://blockstream.info/api/block/" + blockHash + "/txs/"
	var offset int 
	for {
		response, code = HttpGetData(url + strconv.Itoa(offset))

		if code != 200 {
			break
		}
		var transactions []Transaction
		err := json.Unmarshal([]byte(response), &transactions)
		if err != nil {
			panic(err)
		} else {
			for _, transaction := range transactions{
				var inputs []string
				for _, input := range transaction.Vin {
					inputs = append(inputs, input.Txid)
				}
				transactionMap[transaction.Txid] = inputs
			}
		}
		offset += 25
	}
	return transactionMap
}

func BuildAncestorTree(transactionMap map[string][]string) map[string][]string{
	ancestorMap := make(map[string][]string)
	visited := make(map[string]bool)

	for key, _ := range transactionMap {
		visited[key] = false
	}
	

	for key, value := range transactionMap {
		if visited[key]{
			continue
		}
		ancestorMap[key] = []string{}
		for _, id := range value {
			ancestorMap[key] = AppendUnique(ancestorMap[key], Dfs(id, ancestorMap, transactionMap, visited))
		}
		visited[key] = true
	}
	return ancestorMap
}

func Dfs(id string, ancestorMap map[string][]string, transactionMap map[string][]string, visited map[string]bool) []string {
	val, ok := transactionMap[id]
	if !ok {
		return []string{}
	}
	if visited[id] {
		return append([]string{id}, ancestorMap[id]...)
	}
	ancestorMap[id] = []string{}
	for _, txs := range val {
		ancestorMap[id] = AppendUnique(ancestorMap[id], Dfs(txs, ancestorMap, transactionMap, visited))
	}
	visited[id] = true
	return append([]string{id}, ancestorMap[id]...)
}

func AppendUnique(array []string, appendArray[]string) []string {
	for _, val := range appendArray {
		if !Contains(array, val)  {
			array = append(array, val)
		}
	}
	return array
}

func Contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func SortByValAndPrint(ancestorMap map[string][]string, numTransaction int) {
	keys := make([]string, 0, len(ancestorMap))
    for key := range ancestorMap {
        keys = append(keys, key)
    }
    sort.Slice(keys, func(i, j int) bool { return len(ancestorMap[keys[i]]) > len(ancestorMap[keys[j]]) })

    for i:=0; i<numTransaction; i++ {
    	fmt.Println("txid : ",  keys[i] , " , count : ", len(ancestorMap[keys[i]]))
    }
}
