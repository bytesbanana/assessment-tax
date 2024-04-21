FROM golang:1.22-alpine as build-base

WORKDIR /app

COPY go.mod .


RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -v ./...

RUN go build -o ./out/assessment-tax


FROM alpine:3.19.1
COPY --from=build-base /app/out/assessment-tax /app/assessment-tax

WORKDIR /app
EXPOSE 8080
CMD ["/app/assessment-tax"]