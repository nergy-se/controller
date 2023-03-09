
build: export CGO_ENABLED=0
build:
	GOOS=linux GOARCH=amd64 go build -o nergycontroller-linux-amd64 ./cmd/nergycontroller
	GOOS=linux GOARCH=arm GOARM=7 go build -o nergycontroller-linux-arm ./cmd/nergycontroller
