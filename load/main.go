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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/icrowley/fake"
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
		getBooks(host, port)
		isbns := getISBNList(host, port)
		for i, isbn := range isbns {
			getBook(host, port, isbn)
			if i > 20 {
				break
			}
		}
		if rand.Intn(10) > 2 {
			addBook(host, port)
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

func getBooks(host string, port int) *Book {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/books", host, port))
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

func addBook(host string, port int) (*http.Response, error) {
	body, _ := json.Marshal(struct {
		Authors []string
		ISBN    string
		Name    string
		Price   int
	}{
		Authors: []string{fake.FirstName() + " " + fake.LastName()},
	})
	return http.Post(
		fmt.Sprintf("http://%s:%d/books", host, port),
		"application/json",
		ioutil.NopCloser(bytes.NewReader(body)),
	)
}
