FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY src ./src
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/backend ./src/cmd
# The image has no shell: the directory of the documents is made here, so that
# a volume mounted on it belongs to the user the server runs as.
RUN mkdir -p /out/documents

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/backend /backend
COPY --from=build --chown=65532:65532 /out/documents /var/lib/sigta/documents
ENV SIGTA_DOCUMENTS_DIR=/var/lib/sigta/documents
EXPOSE 8080
ENTRYPOINT ["/backend"]
