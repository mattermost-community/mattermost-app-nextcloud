FROM golang:1.18

COPY . /opt/app

# cd /opt/app
WORKDIR /opt/app

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# This container exposes port 8082 to the outside world
EXPOSE 8082

# Run the executable
CMD ["http-server"]