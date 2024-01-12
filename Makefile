
build: export CGO_ENABLED=0
build:
	GOOS=linux GOARCH=amd64 go build -o nergycontroller-linux-amd64 ./cmd/nergycontroller
	GOOS=linux GOARCH=arm GOARM=7 go build -o nergycontroller-linux-arm ./cmd/nergycontroller
	GOOS=linux GOARCH=arm GOARM=7 go build -o modc-linux-arm ./cmd/modc
	GOOS=linux GOARCH=amd64 go build -o modc-linux-amd64 ./cmd/modc

run: 
	gow run ./cmd/nergycontroller

test:
	go test ./...
