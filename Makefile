bin:
	mkdir bin

pgn-linux-x64: bin
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/pgn-linux-x64 pgn.go

pgn-linux-arm64: bin
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/pgn-linux-arm64 pgn.go

pgn-macos-x64: bin
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o bin/pgn-macos-x64 pgn.go

pgn-macos-arm64: bin
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o bin/pgn-macos-arm64 pgn.go

all: pgn-linux-x64 pgn-linux-arm64 pgn-macos-x64 pgn-macos-arm64

clean:
	rm pgn-*
