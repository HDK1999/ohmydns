FROM golang:latest
WORKDIR /project/ohmydns
COPY . .
# 配置依赖镜像源
RUN go env -w GOPROXY=https://goproxy.io,direct;\
    go build -o ./bin/ohmydns ./src
RUN sed -i s@/deb.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list;\
    sed -i s@/security.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list
EXPOSE 53
ENTRYPOINT ["./bin/ohmydns"]
