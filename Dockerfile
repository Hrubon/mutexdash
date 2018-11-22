FROM golang:1.11

ARG path=/go/src/github.com/Hrubon/mutexdash
WORKDIR $path
COPY . .

RUN go build $path

CMD ["./mutexdash -e http://127.0.0.1:2379 -l :8080"]
