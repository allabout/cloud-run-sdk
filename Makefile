.PHONY: test
test:
	bash ./test.sh

.PHONY: test-coverage
test-coverage:
	bash ./test.sh -coverage
