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
	httpClient = &http.Client{}
	isbns := getISBNList()
	for _, isbn := range isbns {
		getBook(isbn)
	}
}

func getISBNList() []string {
	resp, err := http.Get("http://localhost/isbns")
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

func getBook(isbn string) *Book {
	resp, err := http.Get(fmt.Sprintf("http://localhost/books?isbn=%s", isbn))
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
