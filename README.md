set the following environment variables in docker compose:

OpenStack:
- OS_IDENTITYENDPOINT=":5000"
- OS_USERNAME="admin"
- OS_PASSWORD=""
- OS_TENANTID=""
- OS_FOMAINID="default"
- OS_WSOCKHOST=":6080"

minio:
- MINIO_ENDPPOINT=":9001"
- MINIO_ACCESS_KEY=""
- MINIO_SECRET_KEY=""

sample docker-compose file:
```yaml

```