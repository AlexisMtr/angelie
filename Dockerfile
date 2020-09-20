FROM golang:1 as build
COPY . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go mod download && go build -o angelie .

FROM ubuntu as runtime
COPY --from=build /app/angelie /usr/bin/

EXPOSE 10000

ENTRYPOINT ["/usr/bin/angelie"]