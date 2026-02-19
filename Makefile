.PHONY: build test fill tools

build: tools
	./tools/specgen -o refs-openrpc.json \
                -schemas 'src/schemas' \
                -schemas 'src/engine/openrpc/schemas' \
                -methods 'src/eth' \
                -methods 'src/debug' \
                -methods 'src/engine/openrpc/methods'

test: build fill
	./tools/speccheck -v

fill:
	$(MAKE) -C tools fill

tools:
	$(MAKE) -C tools build
