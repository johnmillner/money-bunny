package main

import (
	"github.com/johnmillner/robo-macd/config"
	"github.com/johnmillner/robo-macd/io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClearGraphs(t *testing.T) {
	var files []string

	root := "snapshots"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".css") {
			return nil
		}
		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		log.Printf("removing %s", file)
		_ = os.Remove(file)
	}
}

func TestCancelAllOrders(t *testing.T) {
	config.Config()
	a := io.NewAlpaca()

	for _, order := range a.ListOpenOrders() {
		log.Printf("cancelling order %v", order)
		_ = a.Client.CancelOrder(order.ID)
	}
}
