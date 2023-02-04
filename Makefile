default: install

.PHONY: generate
generate: test
	go generate ./...

.PHONY: install
install:
	go install .

.PHONY: test
test:
	go test -count=1 -parallel=4 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
