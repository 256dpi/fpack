all:
	go fmt .
	go vet .
	golint .
	revive .
