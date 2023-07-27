# MyTunes (Placeholder)

## Prerequisites
Make sure that you have installed the latest version of Go. Visit the official Go website for download instructions.

Make sure you have docker and the docker compose plugin installed. 

This project uses Make. Make sure it's installed on your system. If you're using a Unix-based system like Linux or MacOS, you probably already have it. If you're on Windows, you can get it as part of the MinGW or Cygwin toolchains.

## Installation

### Install Go modules
```
go mod download
```

### Set up env variables
```
cp .env.example .env
```

### Run the project in a dev environment
```
make dev
```

Check project at localhost:3000/

