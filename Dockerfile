# 多阶段:第一阶段编译成静态二进制,第二阶段用极小 alpine 跑
# 用本机现成的 golang:1.23(1.22 alpine 需要拉,慢);build 阶段用 debian based 也没问题
FROM golang:1.23 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/whaleshop ./

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
COPY --from=build /out/whaleshop /whaleshop
EXPOSE 8080
CMD ["/whaleshop"]
