FROM golang:1.27 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY config ./config
COPY scim ./scim

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goscim . && \
    mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/goscim /goscim
COPY --from=build --chown=65532:65532 /out/data /data
COPY config /config

ENV SCIM_CONFIG_DIR=/config \
    SCIM_PORT=:8080

USER nonroot:nonroot
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=15s --retries=3 \
    CMD ["/goscim", "-healthcheck"]

ENTRYPOINT ["/goscim"]
