## Requirements

   - Golang >= 1.12
   - MongoDB >= 4.0 as data store.
   - Redis as a cache store.

## Application Components:

- GraphServer - Contains GraphQL as a backend
- OAuthServer - Complete authentication of application happens here
- MongoDB - Data storage accross the application
- Redis - In-memory cache for the payloads

## Install Go:

 - wget https://storage.googleapis.com/golang/go1.12.linux-amd64.tar.gz
 - sudo rm -rf /usr/local/go
 - sudo tar -C /usr/local -xzf go1.12.linux-amd64.tar.gz
 
## Export Gopath to .bashrc

 - export GOPATH=$HOME/go
 - export PATH=$PATH:$GOPATH/bin
 - export PATH=$PATH:/usr/local/go/bin

## Clone the forked Tribe Platform repository

 - git clone https://github.com/$GITHUB_USERNAME/platform.git

## Setup Environmnet variables

- Use .env.example file to replace the connection strings for MongoDB, Redis e.t.c.

## Start the Graph server

 - go mod tidy
 - go run /cmd/graphserver

## Access Graphql Playground

 - Head to http://localhost:8080 in your browser to play with the current setup

# Using Docker:
## Install Docker CE

- curl -fsSL https://get.docker.com -o get-docker.sh
- sudo sh get-docker.sh
- sudo usermod -aG docker $(whoami)

## Pull and run the Docker Image

- docker pull tribecommerce/platform
- docker run -d -p 8080:8080 --name GraphServer tribecommerce/platform

# Using docker-compose:

## Install docker-compose

- sudo curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
- sudo chmod +x /usr/local/bin/docker-compose

## Run docker-compose.yml to bring up all the stack of tribe platform

- cd platform
- docker-compose up -d

## Destroy docker-compose

- docker-compose down