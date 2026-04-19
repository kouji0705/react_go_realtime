FROM golang:1.26-alpine

WORKDIR /app

# ホットリロードツール「Air」の最新版をインストール
RUN go install github.com/air-verse/air@latest

# コンテナ起動時は air を実行する
CMD ["air"]