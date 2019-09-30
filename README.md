
# Simple HTTP

Simple HTTP app that send GET request to specified service url

## Development

Build
```
go build -o ./bin/simple-http
```

Build docker

```
CGO_ENABLED=0 GOOS=linux go build -o ./bin/simple-http
```

```
docker build -t arifsetiawan/simple-http:0.1 .
```

```
docker run -p 9000:9000 --name simple-http arifsetiawan/simple-http:0.1 -service-url=https://httpbin.org
```
