FROM golang:1.18-alpine3.16 as build

# Set the Current Working Directory inside the container
#WORKDIR $GOPATH/src/downloadsOrganiser
WORKDIR $GOPATH/src/downloadsOrganiser

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

RUN go build -o /downloadsOrganiser ./cmd/downloads_organiser/main.go 


FROM alpine:3.14

COPY --from=build /downloadsOrganiser /downloadsOrganiser

# Run the executable
ENTRYPOINT ["/downloadsOrganiser"]
