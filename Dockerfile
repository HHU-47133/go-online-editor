# 直接基于官方 1.25 镜像，既作为编译环境也作为运行环境
FROM golang:1.25-alpine

# 安装必要的系统工具（git 用于处理子模块，tzdata 用于时区）
RUN apk add --no-cache git tzdata

WORKDIR /app

# 1. 处理依赖（利用 Docker 缓存）
COPY go.* ./
RUN go mod download

# 2. 复制所有源码
COPY . .

# 3. 处理 git 子模块（如果本地源码未包含，则在构建时同步）
RUN if [ -d ".git" ]; then git submodule update --init --recursive; fi

# 4. 编译服务端程序
# 编译后的二进制文件直接留在当前环境即可
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o go-editor main.go

# 设置环境变量
ENV TZ=Asia/Shanghai
EXPOSE 2009

# 运行服务
CMD ["./go-editor"]