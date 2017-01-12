import random
import json
import subprocess
import requests

# script to munge data from isdndb.com into the format used in this example
books = []
for i in range(1, 7):
    print 'Fetching page {} of book results'.format(i)
    resp = requests.get("http://isbndb.com/api/v2/json/JGBAAFCD/books",
                        params={'q': 'accounting', 'p': i}).json()

    for book in resp['data']:
        nb = {
            "name": book["title"],
            "isbn": book["isbn10"],
            "author": [x["name"] for x in book["author_data"]],
            "price": random.randrange(10, 30),
        }
        books.append(nb)

with open('booklist.json', 'w') as output:
    json.dump(books, output)

subprocess.check_call('mongoimport --db bookservice --collection books --drop '
                      '--file ./booklist.json --jsonArray', shell=True)
