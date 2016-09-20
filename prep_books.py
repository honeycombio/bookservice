import random
import json

# script to munge data from isdndb.com into the format used in this example
# populate booklist.* with
# curl -v http://isbndb.com/api/v2/json/JGBAAFCD/books\?q\=accounting > booklist.6

books = []
for blfid in range(1,6):
    with open('booklist.{}'.format(blfid), 'r') as blf:
        bl = json.load(blf)
    for book in bl["data"]:
        nb = {
            "name": book["title"],
            "isbn": book["isbn10"],
            "author": [x["name"] for x in book["author_data"]],
            "price": random.randrange(10, 30),
        }
        books.append(nb)
