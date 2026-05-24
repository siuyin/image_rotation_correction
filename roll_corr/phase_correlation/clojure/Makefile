.PHONY: test help run uberjar

help:
	@echo "test    - Run tests"
	@echo "run     - Run with args, e.g., make run ARGS='-v ~/my.mp4 -i 500'"
	@echo "uberjar - Create uberjar"

test:
	clojure -X:test

run:
	clojure -M -m roll-corr.phase-correlation.core $(ARGS)

uberjar:
	clojure -T:build uber
