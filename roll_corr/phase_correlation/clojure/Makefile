.PHONY: test help

help:
	@echo "test    - Run tests"
	@echo "run     - Run main function (requires reference.png)"

test:
	clojure -X:test

run:
	clojure -M -m roll-corr.phase-correlation.core $(or $(VIDEO), ~/tennis1.mp4) 1000 75
