package main

// generate load for the bookserver

/* tasks:
* get a list of isbns
* get a list of all books
* for each isbn, request the book
* distribute isbn requsets across a normal curve
* post new books
* put to update existing books
* delete books
* ask for nonexistent ISBNs
* delete a nonexistent ISBN
 */

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Book is from the server API
type Book struct {
	ISBN   string   `json:"isbn"`
	Name   string   `json:"name"`
	Author []string `json:"author"`
	Price  int      `json:"price"`
}

var httpClient *http.Client

func main() {
	var host string
	var port int
	flag.StringVar(&host, "host", "localhost", "server host")
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()
	httpClient = &http.Client{}
	for {
		isbns := getISBNList(host, port)
		for _, isbn := range isbns {
			getBook(host, port, isbn)
		}
	}
}

func getISBNList(host string, port int) []string {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/isbns", host, port))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var isbns []string
	json.Unmarshal(body, &isbns)
	return isbns
}

func getBook(host string, port int, isbn string) *Book {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/books?isbn=%s", host, port, isbn))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var b *Book
	json.Unmarshal(body, b)
	return b
}
