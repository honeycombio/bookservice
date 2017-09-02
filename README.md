This is a little demo app of a REST-like API serving book info, backed by a
MongoDB database. Built to illustrate [Honeycomb](https://honeycomb.io/).

To import data, run
```
mongoimport --db bookservice --collection books --drop --file booklist.json --jsonArray --host <mongo_host>
```

To run this app in a kubernetes cluster, run `kubectl apply -f kubernetes/`
