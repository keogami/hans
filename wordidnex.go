package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli/v2"
)

func wordIndexMain(ctx *cli.Context) error {
	if ctx.Args().Len() < 2 {
		return ErrNotEnoughArgs
	}

	inputFilename := ctx.Args().Get(0)
	outputDirname := ctx.Args().Get(1)

	input, err := os.Open(inputFilename)
	if err != nil {
		return err
	}
	defer input.Close()
	bufinput := bufio.NewReader(input)
	wordChan := iterateSequenceList(bufinput, 10)

	output, err := leveldb.OpenFile(outputDirname, nil)
	if err != nil {
		return nil
	}
	defer output.Close()

	total, err := collectWords(wordChan, output)
	if err != nil {
		return err
	}

	fmt.Printf("%d words were indexed", total)

	return nil
}

func iterateSequenceList(input io.Reader, buffer int) <-chan string {
	lineChan := make(chan string, buffer)
	wordChan := make(chan string, buffer)
	wg := new(sync.WaitGroup)
	wg.Add(buffer)
	for i := 0; i < buffer; i++ {
		go splitWords(wg, lineChan, wordChan)
	}

	go scanLine(input, lineChan)

	go func() {
		wg.Wait()
		close(wordChan)
	}()

	return wordChan
}

func collectWords(words <-chan string, output *leveldb.DB) (uint, error) {
	unique := uint(0)
	for word := range words {
		has, err := output.Has([]byte(word), nil)
		if err != nil {
			return unique, err
		}
		if !has {
			val := []byte{0, 0, 0, 0, 0, 0, 0, 1}
			err := output.Put([]byte(word), val, nil)
			if err != nil {
				return unique, err
			}
			unique++
			continue
		}

		value, err := output.Get([]byte(word), nil)
		if err != nil {
			return unique, err
		}
		count := binary.BigEndian.Uint64(value)
		count++
		newval := make([]byte, 8)
		binary.BigEndian.PutUint64(newval, count)

		err = output.Put([]byte(word), newval, nil)
		if err != nil {
			return unique, err
		}
	}

	return unique, nil
}

func splitWords(wg *sync.WaitGroup, lines <-chan string, words chan<- string) {
	defer wg.Done()
	for line := range lines {
		spl := strings.Split(line, " ")
		for _, word := range spl {
			words <- word
		}
	}
}

func scanLine(input io.Reader, outchan chan<- string) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		outchan <- line
	}
	close(outchan)
}
