FROM node:24-alpine AS frontend
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY application/go.mod application/go.sum ./
RUN go mod download
COPY application/*.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /taskboard .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /taskboard /taskboard
COPY --from=frontend /frontend/dist /web
USER 65532:65532
EXPOSE 8080
ENV PORT=8080 WEB_DIR=/web
ENTRYPOINT ["/taskboard"]
