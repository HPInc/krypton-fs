# we assume all test machines have either docker-compose or
# the compose plugin in docker command.
DOCKER_COMPOSE=docker-compose
HAS_DOCKER_COMPOSE := $(shell command -v docker-compose 2> /dev/null)
ifndef HAS_DOCKER_COMPOSE
  DOCKER_COMPOSE=docker compose
  DOCKER_COMPOSE_QUIET_PULL=--quiet-pull
endif

include .env

start: keys
	$(DOCKER_COMPOSE) -p$(PROJECT) up $(DOCKER_COMPOSE_QUIET_PULL) -d

start_fs:
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose.yml -f docker-compose-fs.yml up -d

stop_test:
	-docker rm $(TEST).$(DOMAIN)

stop:
	-$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose.yml -f docker-compose-fs.yml down
	-docker rm $(CLI).$(DOMAIN)

show_files:
	 docker exec $(DB).$(DOMAIN) psql -U $(DB_USER) -d $(DB_NAME) -c "select * from files order by created_at desc limit 1"

test: start_fs
	$(DOCKER_COMPOSE) -p$(PROJECT) \
		-f docker-compose-test.yml up $(DOCKER_COMPOSE_QUIET_PULL) \
		--exit-code-from $(CLI)

clean: stop
	-docker ps -aqf status=exited | xargs docker rm # remove exited
	-rm privateKey.pem publicKey.pem

keys:privateKey.pem publicKey.pem

# do not add this to .PHONY
privateKey.pem:
	openssl genrsa -out privateKey.pem 2048

# do not add this to .PHONY
publicKey.pem:
	openssl rsa -in privateKey.pem -pubout -out publicKey.pem

.PHONY: start stop test test_cli start_fs stop_test show_files clean keys
