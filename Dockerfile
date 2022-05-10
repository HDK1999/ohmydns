FROM golang:latest
WORKDIR /project/ohmydns
COPY . .
# 配置依赖镜像源
RUN go env -w GOPROXY=https://goproxy.io,direct
RUN go build -o ./bin/ohmydns ./src
RUN sed -i s@/deb.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN sed -i s@/security.debian.org/@/mirrors.aliyun.com/@g /etc/apt/sources.list
EXPOSE 53
CMD ["./bin/ohmydns"]
