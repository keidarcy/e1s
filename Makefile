dep:
	go mod download

build: main.go
	go build -o e1s $<

install: main.go
	go install

test:
	go test -v ./...

vet:
	go vet

tag:
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

.PHONY: \
	dep \
	install \
	build \
	vet \
	test