# exporter

## GO Build
```sh
go build -o main .
# Run
./main
```

## Docker
```sh
docker build -t custom-exporter .
# Run
docker run -d --name custom-exporter -p 8080:8080 -e RESOURCE_FILE=/app/test.txt -e RESOURCE_URL=http://URI  -e RESOURCE_JSON=http://URI custom-exporter
```