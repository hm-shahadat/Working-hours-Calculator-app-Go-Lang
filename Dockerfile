FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod .
COPY main.go .
COPY templates ./templates
COPY static ./static
RUN go build -o /workhours .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /workhours /app/workhours
# This is where the JSON data file will live inside the container.
# Attach a persistent disk/volume at /app/data on your host for the
# data to survive restarts and redeploys (see README.md).
ENV DATA_FILE=/app/data/workhours-data.json
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["/app/workhours"]
