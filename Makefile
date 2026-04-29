.PHONY: build run clean

build:
	go build -o imgbed .

run: build
	./imgbed

clean:
	rm -f imgbed imgbed.exe
	rm -f imgbed.db imgbed.db-wal imgbed.db-shm

# Cross compile
linux:
	GOOS=linux GOARCH=amd64 go build -o imgbed-linux .

windows:
	GOOS=windows GOARCH=amd64 go build -o imgbed.exe .

mac:
	GOOS=darwin GOARCH=arm64 go build -o imgbed-mac .
