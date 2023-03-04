FROM golang:latest
RUN mkdir -p /home/go-app
WORKDIR /home/go-app
COPY . .
RUN go mod download
RUN go build -o /go-myapp
RUN ls -al
CMD ["/go-myapp", "server:80"]