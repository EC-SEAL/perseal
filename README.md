# Persistence module for SEAL
## Microservice running on HTTP port 8082
Endpoint routes referenced in [routes.go#L30](routes.go#L30)

## Usage

```bash
docker build -t perseal .
docker run -p 8082:8082 -v config:/config -it perseal
```

## Dependencies
- File `credentials.json` used containing `client_id`/`client_secret` information related to Google Drive oauth client.
- Mock system manager service running on `:8083` as referenced in [sm/sm.go#L58](sm/sm.go#L58)
