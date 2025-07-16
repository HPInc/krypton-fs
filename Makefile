TRIVY_IMAGE:=ghcr.io/aquasecurity/trivy

all: build

build: build_binaries

# Build the binaries for the service.
build_binaries:
	$(GOBUILD) -ldflags $(LDFLAGS) \
	-o $(BIN)/$(TARGET) \
	service/main.go

# Install all files in local folder
install: build
	cp -r config/config.yaml $(BIN)/

# run go tests in container
test: local_test_infra
	make -C tools/compose test

# run integration tests in container
ci_test: local_test_infra docker_image
	make -C tools/compose test

local_test_infra:
	make -C tools/compose

# stop and remove test containers
stop:
	make -C tools/compose stop

# Create a docker image for the service.
docker_image: vendor
	docker build -t $(DOCKER_IMAGE) .

# Publish the fs docker image to Github.
publish: docker push

run: local_test_infra
	FS_DB_SCHEMA=service/db/schema \
	FS_DB_USER=krypton \
	FS_DB_PASSWORD=test \
	FS_CACHE_SERVER=localhost \
	FS_CACHE_PASSWORD=test \
	FS_STORAGE_ENDPOINT=http://localhost:9000 \
	FS_NOTIFICATION_ENDPOINT=http://localhost:9324 \
	FS_SERVER_AUTH_JWKS_URL=http://localhost:9090/api/v1/keys \
	AWS_ACCESS_KEY_ID=minioadmin \
	AWS_SECRET_ACCESS_KEY=minioadmin \
	AWS_REGION=us-east-1 \
	FS_CONFIG_FILE=service/config/config.yaml \
	$(GOCMD) run service/main.go

trivy:
	docker run --rm \
	-v/var/run/docker.sock:/var/run/docker.sock \
	-v"$$HOME/Library/Caches:/root/.cache" \
	$(TRIVY_IMAGE) \
	image -q --severity HIGH,CRITICAL,MEDIUM,LOW --exit-code 1 $(DOCKER_IMAGE)

.PHONY: publish run
.SILENT:

include common.mk
