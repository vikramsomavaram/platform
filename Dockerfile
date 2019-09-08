#Use golang as a base Image
FROM golang:1.13

#Set current working directory
WORKDIR /app

#Install project dependencies
RUN apt-get update && apt-get install -y libssl-dev --no-install-recommends \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

#Copy go mod and sum files
COPY go.mod go.sum ./
#Copy the source from current directory to Working Directory
COPY . .

# Build the Go app
RUN go install ./...

#Expose the port to outside world
EXPOSE 8080

#Run the Executable
ENTRYPOINT ["graphserver"]
