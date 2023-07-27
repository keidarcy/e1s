VERSION := $(shell cat app-version)

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
	echo "Tagging version $(VERSION)"
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)


plan:
	cd tests && terraform plan -var="cluster_count=10" -var="service_count=10" -var="task_count=10"

apply:
	cd tests && terraform apply -var="cluster_count=10" -var="service_count=10" -var="task_count=10"

.PHONY: \
	dep \
	install \
	build \
	vet \
	test