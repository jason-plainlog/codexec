FROM golang:alpine

RUN apk add gcc g++ python3 git make libcap-dev

# install isolate
RUN git clone https://github.com/ioi/isolate.git
WORKDIR isolate
RUN make install

# build server
COPY . /app
WORKDIR /app
RUN go build bin/server.go

CMD ['/app/server']