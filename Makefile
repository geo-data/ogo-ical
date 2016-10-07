##
# Makefile for the Open Groupware iCAL service.
#
# Targets:
# - docker    create a docker image containing Ogo-Ical.
# - clean     delete all generated files.
# - run       run the service.
# - build     compile executable.
#
# Meta targets:
# - all is the default target; it builds the ogo-ical binary.
#

# Golang source files.
SRC_FILES := $(shell ls *.go)

# Build dependencies.
BUILD_DEPS := vendor $(SRC_FILES)

# Create production Docker image components by default.
all: ogo-ical

# Create a docker image for use in production environments.  This first builds
# the development docker image, then copies the ogo-ical binary from this image to
# the current working directory, from where it builds the production image.
docker:
	docker run --rm -v $$(pwd):/tmp/ogo-ical $$(docker build --quiet --file docker/Dockerfile .) cp ogo-ical /tmp/ogo-ical && \
	docker build -t geodata/ogo-ical:latest .

# Create a development environment.
dev:
	docker-compose rm -f -v dev && \
	docker-compose run --rm --service-ports dev

# Run the tests.
test:
	go test

# Remove automatically generated files.
clean:
	@rm -f ogo-ical
	@rm -rf vendor

# Run the service.
run: realize.config.yaml vendor
	realize run

# Build an executable optimised for a linux container environment. See
# <https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07#.otbjvqo3i>.
ogo-ical: OGOICAL_VERSION := $(shell git describe --tags --abbrev=0 --match 'v[0-9]*')
ogo-ical: OGOICAL_COMMIT := $(shell git rev-parse --short HEAD)
ogo-ical: $(BUILD_DEPS)
	CGO_ENABLED=0 \
	GOOS=linux \
	go build -a -tags netgo -ldflags '-w -X main.version=$(OGOICAL_VERSION) -X main.commit=$(OGOICAL_COMMIT)' -o ogo-ical

vendor: glide.yaml glide.lock
	glide install && \
	touch -c vendor

glide.lock:
	glide update && \
	touch -c vendor

glide.yaml:
	glide init

# Targets without filesystem equivalents.
.PHONY: all clean run dev docker
