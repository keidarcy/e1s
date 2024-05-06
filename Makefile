VERSION := $(shell grep -o 'AppVersion = "[^"]*"' internal/utils/info.go | cut -d '"' -f 2)

run:
	go run ./cmd/e1s/main.go

test:
	go test -v ./...

vet:
	go vet

tag:
	echo "Tagging version $(VERSION)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)


plan:
	cd tests && terraform plan -var="cluster_count=10" -var="service_count=10" -var="task_count=1"

apply:
	cd tests && terraform apply -var="cluster_count=10" -var="service_count=10" -var="task_count=1"

.PHONY: \
	dep \
	install \
	build \
	vet \
	test