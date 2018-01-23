package main

import (
	"fmt"
	"time"
	"strconv"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const DataPath string  = "chaindata"
const Proof int = 3

/**
Block structure
 */
type Block struct {
	Index int
	Timestamp time.Time
	LastHash string
	Hash string
	Data string
}

func (s Block) createHeader(nonce int) (string) {
	return strconv.Itoa(s.Index) + s.Timestamp.String() + s.LastHash + s.Data + strconv.Itoa(nonce)
}

func (s Block) createSelfHash() (string) {
	var str string
	for i := 0; i < Proof; i++ {
		str += "0"
	}

	header := s.createHeader(0)

	for i := 0; str != header[0:Proof]; i++ {
		header = s.createHeader(i)
	}

	var hasher = sha256.New()
	hasher.Write([]byte(header))

	var hash = hex.EncodeToString(hasher.Sum(nil))

	return hash
}

func (s Block) writeBlock() {
	f, err := os.Create(DataPath + "/block_" + strconv.Itoa(s.Index) + ".json")
	check(err)

	blockJson, _ := json.Marshal(s)
	f.Write([]byte(blockJson))
}

/**
Chain interface
 */
type Chain interface {
	getBlock(i int) (Block)
	getLastBlock() (Block)
	getChain() ([]Block)
	createBlock(i int, data string, last string) (Block)
}

func getBlock(i int) (Block) {
	data, err := ioutil.ReadFile(DataPath + "/block_" + strconv.Itoa(i) + ".json")
	check(err)

	var b = Block{}
	json.Unmarshal([]byte(data), &b)

	return b
}

func getLastBlock() (Block) {
	files, _ := ioutil.ReadDir(DataPath)
	var blocks = len(files)

	return getBlock(blocks)
}

func getChain() ([]Block) {
	var chain []Block

	files, _ := ioutil.ReadDir(DataPath)
	var blocks = len(files) + 1

	for i := 1; i < blocks; i++ {
		b := getBlock(i)
		chain = append(chain, b)
	}

	return chain
}

func createBlock(data string) (Block) {
	lastBlock := getLastBlock()

	var b = Block{}
	b.Index = lastBlock.Index + 1
	b.Timestamp = time.Now()
	b.LastHash = lastBlock.Hash
	b.Data = data
	b.Hash = b.createSelfHash()
	b.writeBlock()

	return b
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	var chain = getChain()
	w.Header().Set("Content-Length", fmt.Sprint(len(chain)))
	fmt.Fprint(w, chain)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
