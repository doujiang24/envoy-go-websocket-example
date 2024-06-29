.PHONY: build local-build

build:
	docker run --rm -v $(PWD):/go/src/go-filter -w /go/src/go-filter \
		-e GOPROXY=https://goproxy.cn \
		golang:1.21 \
		make local-build

local-build:
	go build -v -o libgolang.so -buildmode=c-shared -buildvcs=false .

run:
	docker run --rm -v $(PWD)/envoy-websocket.yaml:/etc/envoy/envoy.yaml \
		-v $(PWD)/libgolang.so:/etc/envoy/libgolang.so \
		-p 8089:8089 \
		envoyproxy/envoy:contrib-dev \
		envoy -c /etc/envoy/envoy.yaml &

