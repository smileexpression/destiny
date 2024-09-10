# destiny

Something vague.

```bash
docker run -p 9000:9000 -p 9001:9001 \
--name minio \
-d --restart=always \
-e "MINIO_ACCESS_KEY=xxx" \
-e "MINIO_SECRET_KEY=xxx" \
-v /Users/smileexpression/softwareData/minio:/data \
minio/minio server \
/data --console-address ":9001" -address ":9000"
```

```bash
docker run --name redis -p 6379:6379 -d redis
```