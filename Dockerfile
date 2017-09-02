FROM golang

RUN apt-get update
RUN apt-get install -y python wget libsasl2-dev
WORKDIR /go/src/github.com/honeycombio/bookservice
ADD . /go/src/github.com/honeycombio/bookservice
RUN go install ./...

RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 0C49F3730359A14518585931BC711F9BA15703C6 && \
    echo "deb http://repo.mongodb.org/apt/debian jessie/mongodb-org/3.4 main" | tee /etc/apt/sources.list.d/mongodb-org-3.4.list && \
    apt-get update && \
    apt-get install -y mongodb-org
