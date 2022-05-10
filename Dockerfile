FROM golang:latest
WORKDIR /project/ohmydns
COPY . .
# 配置依赖镜像源
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go build -o ./bin/ohmydns ./src
EXPOSE 53
CMD ["./bin/ohmydns"]
