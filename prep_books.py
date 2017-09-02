import random
import json
import subprocess
import sys

# script to munge data from isdndb.com into the format used in this example
if len(sys.argv) > 1:
    mongohost = sys.argv[1]
else:
    mongohost = "localhost"


books = []
for i in range(1, 100):
    print 'Fetching page {} of book results'.format(i)
    subprocess.check_call(
        "wget -q -O booklist.{page} "
        "'http://isbndb.com/api/v2/json/JGBAAFCD/books?q=accounting&p={page}'".format(page=i),
        shell=True)
    with open('./booklist.{}'.format(i)) as f:
        resp = json.load(f)

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
                      '--file ./booklist.json --jsonArray --host %s' % mongohost, shell=True)
