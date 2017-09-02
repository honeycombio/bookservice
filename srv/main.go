package main

// taken from https://github.com/upitau/goinbigdata/tree/master/examples/mongorest
// and the blog post
// http://goinbigdata.com/how-to-build-microservice-with-mongodb-in-golang/

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	flag "github.com/jessevdk/go-flags"
	"goji.io"
	"goji.io/middleware"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DB   = "bookservice"
	COLL = "books"

	QNONE = iota
	QPOOR
	QMED
	QGOOD
	QHEADER = "X-Result-Quality"
)

var qStrings = [4]string{
	"Empty",
	"Poor",
	"Medium",
	"Good",
}

type HandleFunc func(http.ResponseWriter, *http.Request)

func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int, lag float64) {
	delay := rand.NormFloat64()*20 + 100 + lag
	time.Sleep(time.Duration(int(delay)) * time.Millisecond)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

type Book struct {
	ISBN   string   `json:"isbn"`
	Name   string   `json:"name"`
	Author []string `json:"author"`
	Price  int      `json:"price"`
}

type Options struct {
	Mongo string `long:"mongo" default:"localhost"`
}

func main() {
	var options Options
	flagParser := flag.NewParser(&options, flag.PrintErrors)
	_, err := flagParser.Parse()
	if err != nil {
		panic(err)
	}

	session, err := mgo.Dial(options.Mongo)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	logrus.SetFormatter(&logrus.JSONFormatter{})

	mux := goji.NewMux()
	mux.UseC(LogRequest)
	mux.HandleFunc(pat.Get("/books"), getBooks(session))
	mux.HandleFunc(pat.Post("/books"), addBook(session))
	mux.HandleFunc(pat.Put("/books"), updateBook(session))
	mux.HandleFunc(pat.Delete("/books"), deleteBook(session))
	mux.HandleFunc(pat.Get("/isbns"), getISBNs(session))
	http.ListenAndServe("0.0.0.0:8080", mux)
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(DB).C(COLL)

	index := mgo.Index{
		Key:        []string{"isbn"},
		Unique:     true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		log.Fatal(err)
	}
}

func getBooks(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qs := r.URL.Query()
		if len(qs) == 0 {
			allBooks(s)(w, r)
			return
		}
		bookByISBN(s)(w, r)
	}
}
func allBooks(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delay := rand.NormFloat64()*10 + 50
		time.Sleep(time.Duration(int(delay)) * time.Millisecond)

		session := s.Copy()
		defer session.Close()

		c := session.DB(DB).C(COLL)

		var books []Book
		err := c.Find(bson.M{}).All(&books)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all books: ", err)
			return
		}
		respBody, err := json.MarshalIndent(books, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		var lag float64
		if os.Getenv("VERSION") == "2" {
			lag += rand.NormFloat64()*30 + 40
		}

		ResponseWithJSON(w, respBody, http.StatusOK, lag)
	}
}

func addBook(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delay := rand.NormFloat64()*20 + 100
		time.Sleep(time.Duration(int(delay)) * time.Millisecond)
		session := s.Copy()
		defer session.Close()

		var book Book
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&book)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB(DB).C(COLL)

		err = c.Insert(book)
		if err != nil {
			if mgo.IsDup(err) {
				ErrorWithJSON(w, "Book with this ISBN already exists", http.StatusBadRequest)
				return
			}

			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert book: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+book.ISBN)
		w.WriteHeader(http.StatusCreated)
	}
}

func bookByISBN(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		// isbn := pat.Param(ctx, "isbn")
		isbn := r.URL.Query().Get("isbn")

		c := session.DB(DB).C(COLL)

		var book Book
		err := c.Find(bson.M{"isbn": isbn}).One(&book)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed find book: ", err)
			return
		}

		if book.ISBN == "" {
			ErrorWithJSON(w, "Book not found", http.StatusNotFound)
			return
		}

		// judge result quality
		var quality int
		if len(book.Author) != 0 {
			quality++
		}
		if book.Name != "" {
			quality++
		}
		if book.Price != 0 {
			quality++
		}

		w.Header().Set(QHEADER, qStrings[quality])

		respBody, err := json.MarshalIndent(book, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK, 0)
	}
}

func updateBook(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delay := rand.NormFloat64()*20 + 100
		time.Sleep(time.Duration(int(delay)) * time.Millisecond)
		session := s.Copy()
		defer session.Close()

		isbn := r.URL.Query().Get("isbn")

		var book Book
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&book)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB(DB).C(COLL)

		err = c.Update(bson.M{"isbn": isbn}, &book)
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed update book: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Book not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteBook(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delay := rand.NormFloat64()*20 + 100
		time.Sleep(time.Duration(int(delay)) * time.Millisecond)
		session := s.Copy()
		defer session.Close()

		isbn := r.URL.Query().Get("isbn")

		c := session.DB(DB).C(COLL)

		err := c.Remove(bson.M{"isbn": isbn})
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed delete book: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Book not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getISBNs(s *mgo.Session) HandleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB(DB).C(COLL)

		var books []Book
		err := c.Find(bson.M{}).All(&books)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all books: ", err)
			return
		}
		isbns := make([]string, 0)
		for _, book := range books {
			isbns = append(isbns, book.ISBN)
		}
		respBody, err := json.MarshalIndent(isbns, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK, 0)
	}
}

func requestFields(ctx context.Context, r *http.Request, w *ResponseWriterProxy, elapsed int64) map[string]interface{} {
	var remoteAddr string
	addrPort := strings.Split(r.RemoteAddr, ":")
	if len(addrPort) > 0 {
		remoteAddr = addrPort[0]
	}
	fields := map[string]interface{}{
		"method":          r.Method,
		"request":         r.URL.RequestURI(),
		"userAgent":       r.UserAgent(),
		"remoteAddr":      remoteAddr,
		"request_dur_ms":  elapsed,
		"status_code":     w.Status(),
		"response_length": w.Length(),
	}

	for k, v := range r.Header {
		fields["header_"+k] = v
	}

	gojiPattern := middleware.Pattern(ctx)
	if gojiPattern != nil {
		// log our pattern
		fields["gojiPattern"] = gojiPattern.(*pat.Pattern).String()
	}

	return fields
}

// Middleware: log all requests
func LogRequest(handler goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		responseWriter := NewResponseWriterProxy(w)
		handler.ServeHTTPC(ctx, responseWriter, r)
		elapsed := time.Now().Sub(before).Nanoseconds() / 1e6
		logrus.WithFields(requestFields(ctx, r, responseWriter, elapsed)).Info("Handled request")

	})
}

type ResponseWriterProxy struct {
	http.ResponseWriter
	statusCode int
	length     int
}

func NewResponseWriterProxy(inner http.ResponseWriter) *ResponseWriterProxy {
	return &ResponseWriterProxy{inner, 0, 0}
}
func (rw *ResponseWriterProxy) Status() int {
	return rw.statusCode
}
func (rw *ResponseWriterProxy) Length() int {
	return rw.length
}
func (rw *ResponseWriterProxy) Write(bytes []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	rv, err := rw.ResponseWriter.Write(bytes)
	rw.length += rv
	return rv, err
}
func (rw *ResponseWriterProxy) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
