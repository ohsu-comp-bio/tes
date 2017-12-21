
install:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null

proto:
	@protoc \
	  -I ./ \
		-I ./vendor/github.com/googleapis/googleapis \
		--go_out=. tes.proto

tidy:
	@gofmt -w -s *.go

lint:
	@gometalinter --disable-all --enable=vet \
		--enable=golint --enable=gofmt --enable=misspell \
		./...
