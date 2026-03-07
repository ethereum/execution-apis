.PHONY: build test fill tools

SPECFLAGS := -schemas 'src/schemas' \
	-schemas 'src/engine/openrpc/schemas' \
	-methods 'src/eth' \
	-methods 'src/debug' \
	-methods 'src/txpool' \
	-methods 'src/engine/openrpc/methods'


build: tools
	./tools/specgen -o refs-openrpc.json $(SPECFLAGS)
	./tools/specgen -o openrpc.json -deref $(SPECFLAGS)

test: tools
	./tools/speccheck -v

fill:
	$(MAKE) -C tools fill

tools:
	$(MAKE) -C tools build
