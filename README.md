# v2ray-core

## Docker Build

```bash
docker build -t link-server .
```

## Docker Run

```bash
docker run -d \
  --name link-server-01 \
  --restart always \
  -p 8888:8080 \
  -e V2RAY_REMOTE_USER_LIST_API="https://link.dmumedia.com/api/users/list" \
  link-server:latest
```
