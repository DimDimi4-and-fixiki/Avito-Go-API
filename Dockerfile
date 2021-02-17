FROM golang:latest

ENV GOPATH=/

COPY ./ ./


RUN go get github.com/gorilla/mux
RUN go get go.mongodb.org/mongo-driver/bson
RUN go get go.mongodb.org/mongo-driver/bson/primitive
RUN go get go.mongodb.org/mongo-driver/mongo
RUN go get github.com/go-playground/validator/v10

RUN go mod download
RUN go build -o api ./main.go


CMD ["./api"]

