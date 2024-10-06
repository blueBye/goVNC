#FROM registry.nmcr.ir/apps/golang:1.22-alpine3.19
FROM golang:1.22-alpine3.19

WORKDIR /app

# Install ffmpeg (Yeah this needs both binary and installation)
RUN apk add ffmpeg

# Copy binaries (make sure these files are executable by running chmod +x)
COPY ffmpeg/ ./
COPY vnc2video ./

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Set Go environment variables
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn

# Download Go modules
RUN go mod download

# Copy the source code and required files
COPY *.go ./

# Build
RUN go build -o vnc_recorder

EXPOSE 8080

CMD ["/app/vnc_recorder"]
