ANTLR_VERSION ?= 4.13.2
ANTLR_JAR     ?= build/antlr-$(ANTLR_VERSION)-complete.jar
ANTLR_URL     ?= https://www.antlr.org/download/antlr-$(ANTLR_VERSION)-complete.jar
JAVA_IMAGE    ?= eclipse-temurin:21-jdk

.PHONY: all build vet test race cover check generate clean

all: check

build:
	go build ./...

# The unreachable-code analyser is off project-wide, which is not what we would
# choose. The ANTLR Go target emits unreachable statements in the generated
# parser, and vet surfaces a dependency's diagnostics in every package that
# imports it, so the check cannot be disabled for the generated package alone.
# Every other analyser stays on. Revisit if the ANTLR generator stops emitting
# them.
vet:
	go vet -unreachable=false ./...

test:
	go test ./... -count=1

race:
	go test ./... -count=1 -race

# Runs the suite against a real Couchbase started in a container. Slow: the
# image is large and the cluster takes a couple of minutes to become healthy,
# which is why it is behind a build tag rather than part of `make test`.
integration:
	go test -tags integration -timeout 20m -count=1 -v ./test/integration

cover:
	@mkdir -p build
	go test ./... -count=1 -coverprofile=build/coverage.out
	go tool cover -func=build/coverage.out | tail -1

check: build vet race

# Regenerate the SCIM filter parser from the grammar.
#
# ANTLR is a Java tool, so it runs in a container rather than requiring a JDK on
# the machine. The jar is downloaded on demand instead of being committed.
#
# scim/parser/scimfilter_listener_implement.go and errors.go are hand-written
# and must not be overwritten; the generator only produces the four
# scimfilter_{lexer,parser,listener,base_listener}.go files and the .tokens.
generate: $(ANTLR_JAR)
	@mkdir -p build/antlr-out
	cp ScimFilter.g4 build/
	docker run --rm -v "$(PWD)/build":/work -w /work $(JAVA_IMAGE) \
		java -jar $(notdir $(ANTLR_JAR)) -Dlanguage=Go -o antlr-out ScimFilter.g4
	cp build/antlr-out/scimfilter_base_listener.go \
	   build/antlr-out/scimfilter_lexer.go \
	   build/antlr-out/scimfilter_listener.go \
	   build/antlr-out/scimfilter_parser.go \
	   build/antlr-out/ScimFilter.tokens \
	   build/antlr-out/ScimFilterLexer.tokens \
	   scim/parser/
	gofmt -w scim/parser
	@# The generated copies must not linger: ./... would pick them up as a
	@# second package with the same contents.
	rm -rf build/antlr-out build/ScimFilter.g4
	@echo "Parser regenerated. Run 'make check'."

$(ANTLR_JAR):
	@mkdir -p build
	curl -sSL -o $@ $(ANTLR_URL)

clean:
	rm -rf build
