TARGETS_NOVENDOR=$(shell glide novendor)

mackerel-plugin-nginx-cache: *.go
	go build -o $@

bundle:
	glide install

check:
	go test $(TARGETS_NOVENDOR)

fmt:
	@echo $(TARGETS_NOVENDOR) | xargs go fmt

clean:
	rm -rf mackerel-plugin-nginx-cache
