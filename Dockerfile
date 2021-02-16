FROM golang:latest

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go get github.com/gorilla/mux
RUN go get go.mongodb.org/mongo-driver/bson
RUN go get go.mongodb.org/mongo-driver/bson/primitive
RUN go get go.mongodb.org/mongo-driver/mongo
RUN go get gopkg.in/go-playground/validator.v10
RUN go build

CMD ["./Avito-Go-API"]

