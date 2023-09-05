# 使用官方的Go映像作為基礎映像
FROM golang:1.19

# 設置工作目錄
WORKDIR /app

# 將Go文件複製到容器中
COPY go /app

# 編譯Go程序
RUN cd /app && go build -o main .

# 指定容器啟動時要運行的命令
CMD ["./main"]