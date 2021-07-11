.PHONY: test
test:
	bash ./test.sh

.PHONY: test-coverage
test-coverage:
	bash ./test.sh -coverage

.PHONY: godoc
godoc:
	godoc -http=:6060&
	sleep 1
	open http://localhost:6060/pkg/github.com/ishii1648/cloud-run-sdk/

.PHONY: stop-godoc
stop-godoc:
	kill $$(ps aux | grep godoc | grep -v grep  | awk '{print $$2}')