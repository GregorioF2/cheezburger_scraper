# Compile stage
FROM golang:1.16.5 AS build-env

RUN apt update && apt -y upgrade 

RUN apt -y install chromium

WORKDIR /app


ADD ./ ./

RUN go mod download

RUN go build -o /docker-gs-ping

CMD [ "/docker-gs-ping" ]