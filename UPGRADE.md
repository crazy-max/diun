## 1.3.0 > 1.4.0

As the container runs as a non-root user, you have to first stop the container and change permissions to `data` volume:

```
docker-compose stop
chown -R 1000:1000 data/
docker-compose pull
docker-compose up -d
```
