FROM golang:alpine
RUN apk add git
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .
EXPOSE 1234
CMD ["/app/main"]