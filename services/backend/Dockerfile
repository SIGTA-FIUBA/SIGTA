FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY src ./src
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/backend ./src/cmd

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/backend /backend
EXPOSE 8080
ENTRYPOINT ["/backend"]
