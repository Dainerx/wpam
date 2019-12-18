# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Binary names
BINARY_NAME=wpam
BINARY_UNIX=$(BINARY_NAME)_unix

# Make commands
all: test build
build: 
		$(GOBUILD) -o $(BINARY_NAME) -v
test: 
		$(GOTEST) -v -timeout 90s ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		$(GOBUILD) -o $(BINARY_NAME) -v
		./$(BINARY_NAME)
help: Makefile
	@echo "Choose a command run in "$(PROJECTNAME)":"
	@echo "make all: test and build the project."
	@echo "make build: build the project."
	@echo "make test: run all tests."
	@echo "make clean: clean the project."
	@echo "make run: build and run the project."

