package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mic6090/fsarch"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"
)

var config fsarch.Agent

func main() {
	startTime := time.Now()
	defer func() {
		fmt.Println(time.Now().Sub(startTime))
	}()

	config = fsarch.LoadConfig().Agent

	/*	if len(os.Args) < 2 {
			log.Fatal("No arguments")
		}
		datapath := os.Args[1]
	*/
	//tsFile := "./timestamp"
	tsTime := time.Time{}
	finfo, err := os.Stat(config.Tsfile)
	if err == nil {
		tsTime = finfo.ModTime()
	}
	_ = fsarch.Touch(config.Tsfile)

	f, err := os.Open(config.Storage)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	totalCount := 0
	var totalSize int64 = 0

	for {
		finfos, err := f.Readdir(1024)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			break
		}
		for _, item := range finfos {
			if tsTime.Sub(item.ModTime()) < 0 &&
				item.Mode().IsRegular() &&
				fsarch.CheckName(item.Name()) {
				if postFile(path.Join(config.Storage, item.Name())) == http.StatusOK {
					totalCount++
					totalSize += item.Size()
					if totalCount%10 == 0 {
						fmt.Printf("Files sent: %d, total size: %d...\r", totalCount, totalSize)
					}
				}
			}

		}
	}
	log.Printf("Done. Files sent: %d, total size: %d\r", totalCount, totalSize)
}

func postFile(fileName string) int {
	var b bytes.Buffer
	h, err := os.Open(fileName)
	if err != nil {
		log.Print(err)
		return -1
	}
	size, err := io.Copy(&b, h)
	fmt.Printf("File: %s, bytes read: %d\n", fileName, size)
	if err != nil {
		log.Print(err)
		return -1
	}
	_ = h.Close()
	hash := sha256.Sum256(b.Bytes())
	//fmt.Printf("sha256: \"%s\"\n", hex.EncodeToString(hash[:]))

	var request bytes.Buffer
	w := multipart.NewWriter(&request)
	fw, err := w.CreateFormFile("file", path.Base(path.Clean(fileName)))
	if err != nil {
		log.Fatal(err)
	}
	if _, err = io.Copy(fw, bytes.NewReader(b.Bytes())); err != nil {
		log.Fatal(err)
	}
	_ = w.WriteField("hash", hex.EncodeToString(hash[:]))
	_ = w.Close()

	req, err := http.NewRequest("POST", config.Server, &request)
	if err != nil {
		return -1
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	return resp.StatusCode
}
