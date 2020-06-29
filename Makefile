build:
	GOOS=linux GOARCH=amd64 go build -o bin/djinni-renderer_linux_amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/djinni-renderer_darwin_amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/djinni-renderer_windows_amd64 main.go

run:
	go run main.go examples/optionallist.djinni examples/djinni.template

test:
	go run main.go examples/optionallist.djinni examples/djinni.template > /tmp/optionallist.djinni
	diff --ignore-all-space --ignore-blank-lines /tmp/optionallist.djinni examples/optionallist.djinni

dependencies:
	go list -m all
