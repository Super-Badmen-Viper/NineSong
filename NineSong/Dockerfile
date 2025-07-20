#Build ninesong Model
# 第一阶段：构建
FROM golang:1.24.0-alpine AS builder
WORKDIR /app
# 1. 优先复制依赖文件
COPY go.mod go.sum ./
# 2. 设置国内镜像代理加速下载
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=1
# 3. 安装编译环境
RUN apk add --no-cache g++ make git musl-dev
# 4. 预下载依赖
RUN go mod download
# 5. 克隆并准备词典文件
RUN git clone https://github.com/yanyiwu/cppjieba.git && \
    mkdir -p /app/jieba-dict && \
    cp -r cppjieba/dict/* /app/jieba-dict
# 6. 复制源码
COPY . .
# 7. 静态编译
RUN go build -ldflags="-extldflags '-static'" -o main cmd/main.go
# 第二阶段：运行环境
FROM alpine:latest
WORKDIR /app
# 8. 配置阿里云镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
# 9. 安装运行时依赖
RUN apk add -U --no-cache ffmpeg libc6-compat libstdc++
# 10. 复制构建产物和词典文件
COPY --from=builder /app/main /app/main
COPY --from=builder /app/.env /app/.env
COPY --from=builder /app/jieba-dict /app/jieba-dict
# 11. 安全加固
RUN adduser -D appuser && \
    chown -R appuser:appuser /app && \
    chmod -R 755 /app/jieba-dict
USER appuser
# 12. 设置环境变量
ENV JIEBA_DICT_PATH=/app/jieba-dict
CMD ["/app/main"]

#FROM golang:1.24.0 AS builder
#WORKDIR /app
#ADD . /app
#RUN go mod download
#RUN go build -o main cmd/main.go
#FROM debian:bookworm-slim
#WORKDIR /app
#COPY --from=builder /app/main /app/main
#COPY --from=builder /app/.env /app/.env
#CMD ["/app/main"]

## 构建阶段
#FROM golang:1.24.0-alpine AS builder
#WORKDIR /app
#COPY . .
#RUN go mod download
#RUN go build -o main cmd/main.go
#
## 运行阶段
#FROM alpine:latest
#
## 安装FFmpeg及依赖[3,6](@ref)
#RUN apk update && \
#    apk add --no-cache \
#    ffmpeg \
#    libc6-compat && \
#    rm -rf /var/cache/apk/*
#
## 设置应用运行环境
#WORKDIR /app
#COPY --from=builder /app/main /app/main
#COPY --from=builder /app/.env /app/.env
#
## 验证FFmpeg安装[6](@ref)
#RUN ffmpeg -version
#
## 启动应用
#CMD ["/app/main"]