DIST_DIR=./dist
CMD_DIR=./cmd

SLICES = $(shell ls cmd)

.PHONY: clean
clean:
	rm -rf dist/

.PHONY: all
all: $(SLICES)

PHONY: $(SLICES)
$(SLICES):
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$@/bootstrap $(CMD_DIR)/$@/*.go \
	&& cd $(DIST_DIR)/$@/ \
	&& chmod +x bootstrap	

.PHONY: run-lb
run-lb:
	$(DIST_DIR)/loadbalancer/bootstrap