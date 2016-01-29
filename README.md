# Readme

This project can be used to mirror an incoming webhook to N other endpoints.
This is originally used to work around the fact that Helpscout only allows
setting up 1 webhook endpoint, while we want to send it to more locations.

Webhooks get retried up to 10 times if the connection failed totally, but a 500
return response for example will not trigger this behavior.

There is a catch, this will always return a 200 to the original webhook
request. No proxying of requests will be done, it's just fire and forget. So if
you are dependent on hooks being retried on a non-20x code, *consider yourself
warned* ;-)

# How to run

```bash
go build -a .
FORWARDHOOK_SITES="http://127.0.0.1:4567,http://127.0.0.1:4568" ./forwardhook
```

### How to build the docker container

This is based on the [minimal docker container](http://blog.codeship.com/building-minimal-docker-containers-for-go-applications/) article from Codeship.

SSL certificates are bundled in to get around x509 errors when requesting SSL
endpoints.

```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o forwardhook .
docker build -t bittersweet/forwardhook -f Dockerfile.scratch .

# Push to docker hub
docker push bittersweet/forwardhook
```

## Run it locally

```bash
docker run -e "FORWARDHOOK_SITES=https://site:port/path" --rm -p 8000:8000 -it bittersweet/forwardhook
curl local.docker:8000
```
