FROM golang:1.19

WORKDIR /app

COPY . .

RUN go build -o filestorage

EXPOSE 8080

CMD ["./filestorage"]