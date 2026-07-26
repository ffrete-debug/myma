# Stage 1: Build Go binary
FROM golang:1.24-alpine AS builder-go
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
COPY server/ .
RUN go build -o /app/bin/ark-commander .

# Stage 2: Build Next.js frontend
FROM node:20-alpine AS builder-ui
WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm install
COPY ui/ ./
RUN npm run build

# Stage 3: Final production image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache nodejs
COPY --from=builder-go /app/bin/ark-commander /app/bin/ark-commander
COPY --from=builder-ui /app/ui/.next/standalone ./
COPY --from=builder-ui /app/ui/.next/static ./.next/static
COPY --from=builder-ui /app/ui/public ./public
COPY --from=builder-ui /app/ui/messages ./messages
RUN mkdir -p /data
EXPOSE 8080 3000
ENV DB_PATH=/data/ark-commander.db
ENV SERVER_PORT=8080
ENV NEXT_PUBLIC_API_BASE=http://localhost:8080/api
CMD ["sh", "-c", "/app/bin/ark-commander & node /app/server.js & wait"]
