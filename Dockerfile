FROM golang:latest

WORKDIR /app

RUN echo "Europe/Chisinau" > /etc/timezone
RUN dpkg-reconfigure -f noninteractive tzdata
ENV TZ=Europe/Chisinau

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

LABEL org.opencontainers.image.source https://github.com/dmitriy-robu/go-outpost

CMD ["air", "-c", ".air.toml"]