FROM golang:latest
WORKDIR /project/ohmydns
COPY . .

RUN go build -o ./bin/ohmydns ./src
EXPOSE 53
CMD ["./bin/ohmydns"]
