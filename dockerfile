FROM golang:1.15 as modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

FROM golang:1.15 as builder

COPY --from=modules /go/pkg /go/pkg

# Добавляем непривилегированного пользователя
RUN useradd -u 10001 dupl

RUN mkdir -p /src
ADD . /src
WORKDIR /src

# Собираем бинарный файл
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o /duplicates ./cmd/duplicates

# Final stage: Run the binary
FROM scratch

# Не забываем скопировать /etc/passwd с предыдущего стейджа
COPY --from=builder /etc/passwd /etc/passwd
USER dupl

COPY --from=builder /duplicates /duplicates

CMD ["/duplicates"]