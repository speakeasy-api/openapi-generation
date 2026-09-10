.PHONY: *

docs:
	rm -rf src || true
	speakeasy generate docs --schema ./tests/specs/review.yaml --out ./ --langs go,typescript --compile

build:
	pnpm run build

run-server:
	$(MAKE) build
	go run server.go