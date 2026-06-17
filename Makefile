.PHONY: build test fill tools lint

SPECFLAGS := -schemas 'src/schemas' \
	-schemas 'src/engine/openrpc/schemas' \
	-methods 'src/eth' \
	-methods 'src/debug' \
	-methods 'src/txpool' \
	-methods 'src/engine/openrpc/methods' \
	-methods 'src/testing' \
	-error-groups 'src/error-groups'


build: tools
	./tools/specgen -o refs-openrpc.json $(SPECFLAGS)
	./tools/specgen -o openrpc.json -deref $(SPECFLAGS)

test: tools
	./tools/speccheck -v

lint:
	@[ -f refs-openrpc.json ] || $(MAKE) build >/dev/null
	cd tools && go tool openrpc-linter lint ../refs-openrpc.json -r ../openrpc-lint.yml

fill:
	$(MAKE) -C tools fill

tools:
	$(MAKE) -C tools build
