FROM ghcr.io/hpinc/krypton/krypton-go-builder AS builder

ADD . /go/src/fs
WORKDIR /go/src/fs

# build the source
RUN make build_binaries

# use a minimal alpine image for services
FROM ghcr.io/hpinc/krypton/krypton-go-base

# set working directory
WORKDIR /go/bin

COPY --from=builder /go/src/fs/bin .
COPY --from=builder /go/src/fs/service/config/config.yaml .
COPY --from=builder /go/src/fs/service/db/schema /go/bin/schema/

USER 1001
EXPOSE 8989/tcp

# run the binary
ENTRYPOINT ["/go/bin/fs"]
