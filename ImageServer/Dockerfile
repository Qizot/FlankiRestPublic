ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache  git

WORKDIR /src

COPY ./go.mod  ./
RUN go mod download


COPY . .

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /image_service .

FROM scratch AS final

COPY --from=builder /image_service /image_service

EXPOSE 5555

ENTRYPOINT ["/image_service"]