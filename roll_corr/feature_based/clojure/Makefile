.PHONY: test help

help:
	@echo "test    - Run tests"
	@echo "run     - Run main function (requires reference.png)"

test:
	clojure -X:test

run:
	clojure -M -m roll-corr.feature-based.core reference.png 1000 75
