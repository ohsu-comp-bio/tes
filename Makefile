
proto:
	@protoc tes.proto

tidy:
	@gofmt -w -s ./...

lint:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@gometalinter --disable-all --enable=vet \
		--enable=golint --enable=gofmt --enable=misspell \
		./...
