FROM golang:1.25-alpine AS build
WORKDIR /src
COPY application/go.mod ./
COPY application/*.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /taskboard .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /taskboard /taskboard
USER 65532:65532
EXPOSE 8080
ENV PORT=8080 DATA_FILE=/data/tasks.json
ENTRYPOINT ["/taskboard"]
