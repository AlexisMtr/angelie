FROM golang:1.15.2 as builder
ARG BUILD_VERSION
COPY . /app
WORKDIR /app
ENV GO111MODULE=on
ENV GOPROXY=direct
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.Version=${BUILD_VERSION}" -v -o angelie .

FROM alpine:latest
EXPOSE 10000
COPY --from=builder /app/angelie /bin
ENTRYPOINT [ "/bin/angelie" ]