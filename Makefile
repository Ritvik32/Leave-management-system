
build:
	go build -o build/

.PHONY: build


start:
	go run main.go

docker:
	docker build -t leaves .
	sudo docker run -p 1234:1234
	







