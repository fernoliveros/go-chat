# Stage 1: Build the Angular application
FROM --platform=linux/amd64 node:lts-alpine AS ngbuilder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build --configuration=production

# Stage 2: Build the Go app
FROM --platform=linux/amd64 golang:1.25.4 AS gobuilder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o /app/backend/gochat .


# Stage 3: Runtime
FROM scratch
WORKDIR /app
COPY --from=ngbuilder /app/dist /dist
COPY --from=gobuilder /app/backend/gochat /gochat
EXPOSE 8080
CMD ["/gochat"]