build:
	dep ensure
	env GOOS=linux go build -ldflags="-s -w" -o bin/weather weather/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/airq airq/main.go