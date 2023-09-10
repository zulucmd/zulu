TESTPROG_DIR := $(CURDIR)/testprog

TESTS := $(patsubst tests/Dockerfile.%,%,$(wildcard tests/Dockerfile.*))
SHELLS := bash fish

containing = $(foreach v,$2,$(if $(findstring $1,$v),$v))
not-containing = $(foreach v,$2,$(if $(findstring $1,$v),,$v))

.PHONY: all
all: $(TESTS)

.PHONY: build-linux
build-linux:
	cd $(TESTPROG_DIR) && make build-linux

.PHONE: $(TESTS)
$(TESTS): build-linux
	src/test-all.sh $@

.PHONY: $(SHELLS)
$(SHELLS):
	$(MAKE) $(call containing,$@,$(TESTS))

.PHONY: test
test: clean
	@echo "NOT READY"

.PHONY: macos
macos: mac

PHONY: mac
mac:
	cd $(TESTPROG_DIR) && make clean
	cd $(TESTPROG_DIR) && make
	src/test-completion.sh bash
	src/test-completion.sh fish

.PHONY: clean
clean:
	cd $(TESTPROG_DIR) && make clean
