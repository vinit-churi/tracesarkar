# The collector as a container, for the Cloud Run job in asia-south1.
#
# Cloud Run has no disk and no machine to maintain, which is why it is worth
# testing against the VM in ADR 0015. The open question it answers: does a job's
# outbound traffic present as an Indian address, which is all BMC will answer.

FROM golang:1.26 AS build
WORKDIR /src

# Dependencies first, so edits to the code do not re-download them.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/ingest ./cmd/ingest

# distroless/static carries the CA bundle and nothing else: no shell, no package
# manager, nothing to patch.
FROM gcr.io/distroless/static-debian12
WORKDIR /app

COPY --from=build /out/ingest /app/ingest
# The source register is configuration the collector reads at runtime. Baking it
# into the image keeps it in lockstep with the binary, exactly as shipping it
# beside the binary does for the VM.
COPY data/sources.yaml /app/sources.yaml

ENTRYPOINT ["/app/ingest"]
# One nightly invocation does both collectors, so the schedule needs no argument
# overrides and the trigger identity needs nothing beyond run.invoker.
CMD ["all", "--register", "/app/sources.yaml"]
