//go:generate ./generate.sh

package main

import (
	"github.com/dzeckelev/uploader"
	"path/filepath"
	"os"
	"time"
)

var (
	crt = filepath.Join(os.Getenv("GOPATH"), "src", "github.com",
		"dzeckelev", "uploader", "example", "testdata", "mycert.crt")

	key = filepath.Join(os.Getenv("GOPATH"), "src", "github.com",
		"dzeckelev", "uploader", "example", "testdata", "mykey.key")

	srvPath = filepath.Join(os.Getenv("GOPATH"), "src", "github.com",
		"dzeckelev", "uploader", "example", "out")

	fileIn = filepath.Join(os.Getenv("GOPATH"), "src", "github.com",
		"dzeckelev", "uploader", "example", "testdata", "file.txt")

	address = "localhost:8889"
)

func srv() {
	s := uploader.NewServer(address, srvPath, crt, key)

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func cli(timeout time.Duration) {
	time.Sleep(timeout)

	cli := uploader.NewClient(address)
	defer cli.Close()

	if err := cli.Dial(time.Second); err != nil {
		panic(err)
	}

	pathID, err := cli.CreatePath()
	if err != nil {
		panic(err)
	}

	if err := cli.Upload(fileIn, pathID); err != nil {
		panic(err)
	}

	fileInStat, err := os.Stat(fileIn)
	if err != nil {
		panic(err)
	}

	fileOutStat, err := os.Stat(filepath.Join(srvPath, pathID, "file.txt"))
	if err != nil {
		panic(err)
	}

	if fileOutStat.Size() != fileInStat.Size() {
		panic("not equal file sizes")
	}

	os.RemoveAll(srvPath)
	os.MkdirAll(srvPath, os.ModePerm)
}

func main() {
	go srv()

	cli(time.Second * 3)
}
