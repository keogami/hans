package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli/v2"
)

// KVPair encaps the keys/value pair
type KVPair struct {
	Key   string
	Value []byte
}

// JsonKVPair is just a string to string map corresponding to the Key/Value
type JsonKVPair map[string]string

// genDBMain loads the input file (first arg) and pushes each key-value pair in the input
// to a KVS DB like (LevelDB). It tries to as efficient as possible
func genDBMain(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		return ErrNotEnoughArgs
	}

	input, output := ctx.Args().Get(0), ctx.Args().Get(1)

	inputFile, err := os.Open(input)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	db, err := leveldb.OpenFile(output, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	const iterBuffer int = 10
	iterChan := iterateKVJson(inputFile, iterBuffer)

	n, err := collectKVPair(iterChan, db)
	if err != nil {
		return err
	}

	fmt.Printf("%d entries stored in %s\n", n, output)
	return nil
}

// iterateKVJson parses the file and sends each parsed key/value pair out throuhg the iterChan
func iterateKVJson(input io.Reader, buffer int) <-chan KVPair {
	iterChan := make(chan KVPair, buffer)
	go func() {
		defer close(iterChan)

		data := make(map[string]string)
		if err := json.NewDecoder(input).Decode(&data); err != nil {
			// TODO(keogami): add logging
			return
		}

		for key, value := range data {
			iterChan <- KVPair{
				Key: key, Value: []byte(value),
			}
		}
	}()
	return iterChan
}

// collectKVPair collects the key/value from pairs into the output writer
func collectKVPair(pairs <-chan KVPair, output *leveldb.DB) (uint, error) {
	var count uint = 0
	for pair := range pairs {
		err := output.Put([]byte(pair.Key), pair.Value, nil)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
