# syntax=docker/dockerfile:1

# Stage 1

FROM golang:1.18-alpine AS build

# create a working directory inside the image
WORKDIR /app

# copy Go modules and deps to image
COPY go.mod ./

# download Go modules and deps
RUN go mod download

# copy all files ending with .go
COPY . .

# Prevent dynamically linking to C lib
ENV CGO_ENABLED=0

# compile application
RUN go build -o /go/bin/gargoyle .

###

# Stage 2

FROM scratch

WORKDIR /

COPY config.json config.json

COPY --from=build /go/bin/gargoyle /bin/gargoyle

ENTRYPOINT ["./bin/gargoyle"]