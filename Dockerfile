FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

ADD ../../Desktop/tinkoff-zookeeper-hw/2024-spring-ab-go-hw-2-DimaGitHahahab /app

RUN CGO_ENABLED=0 GOOS=linux go build -o build/election cmd/election/main.go

FROM alpine:latest

COPY --from=build /app/build/* /opt/

ENTRYPOINT [ "/opt/election" ]
CMD [ "run", "--zk-servers", "zoo1:2181,zoo2:2182,zoo3:2183", "--leader-timeout", "5s", "--attempter-timeout", "5s", "--file-dir", "/tmp", "--storage-capacity", "10" ]

