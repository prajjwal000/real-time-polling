FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
