# Build docker image

```
docker buildx build --platform linux/arm64 -t pi-gochat .
```

# Load docker onto raspberry pi

```
docker save pi-gochat | ssh <ssh_user>@<pi_ip_address> 'docker load'
```

# Run container in raspberry pi

```
ssh <ssh_user>@<pi_ip_address>

docker run --name gochat -p 8080:8080 pi-gochat
```