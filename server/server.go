package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mic6090/fsarch"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var config fsarch.Server

func main() {
	config = fsarch.LoadConfig().Server

	for _, item := range []string{config.Backup, config.Datapath, config.Hashpath} {
		fileinfo, err := os.Stat(item)
		if err != nil {
			makeDirs()
			break
		}
		if !fileinfo.Mode().IsDir() {
			log.Fatalf("Path element %s is not directory!", item)
		}
	}

	http.HandleFunc("/", postHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.Bind, config.Port), nil))
}

func makeDirs() {
	for i := 0; i < 16; i++ {
		l1 := fmt.Sprintf("%x", i)
		for k := 0; k < 16; k++ {
			l2 := fmt.Sprintf("%x", k)
			err := os.MkdirAll(path.Join(config.Datapath, l1, l2), 0755)
			if err != nil {
				log.Fatal(err)
			}
			err = os.MkdirAll(path.Join(config.Hashpath, l1, l2), 0755)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	err := r.ParseMultipartForm(8 << 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hash, ok := r.MultipartForm.Value["hash"]
	if !ok {
		log.Println("No hash data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, ok := r.MultipartForm.File["file"]
	if !ok {
		log.Println("No file data")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var f bytes.Buffer
	mh, _ := file[0].Open()
	defer mh.Close()
	io.Copy(&f, mh)
	localHash := sha256.Sum256(f.Bytes())
	h256 := hex.EncodeToString(localHash[:])
	if hash[0] != h256 {
		log.Println("Data hash mismatch")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hashName := path.Join(config.Hashpath, string(h256[0]), string(h256[1]), h256)

	fname := (*file[0]).Filename
	if !fsarch.CheckName(fname) {
		log.Println("File name check error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fullName := path.Join(config.Datapath, string(fname[0]), string(fname[1]), fname)
	fmt.Println(fullName)

	// check for hash name existence
	_, err = os.Stat(hashName)
	if err != nil {
		// write new hash object
		hh, err := os.Create(hashName)
		if err != nil {
			log.Println("Hash file create error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.Copy(hh, bytes.NewReader(f.Bytes()))
		hh.Close()
	} else {
		// read existing object and check it with received
		hh, err := os.Open(hashName)
		if err != nil {
			log.Println("Hash file open error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer hh.Close()
		var hObj bytes.Buffer
		io.Copy(&hObj, hh)
		if bytes.Compare(f.Bytes(), hObj.Bytes()) != 0 {
			log.Printf("File content mismatch: file %s, hash %s\n", fname, h256)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = os.Link(hashName, fullName)
	if err != nil {
		log.Println("Link create error: ", err)
	}
}
