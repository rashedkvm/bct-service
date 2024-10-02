# bct-service
BCT Service examples

## APP_PORT 
Default to 8080

## heatlhz endpoint

Request: 
```console
curl bct-service.bct-service.svc.cluster.local:8080/api/healthz
```

Response: 
```console
{"etag":"8a906761-6ce4-4f7a-a981-6100dd6b93d0","status":"ok:","time":"2024-10-01T21:25:45.031615-04:00"}
```

## upload endpoint (uploads a file to http server)

Request: 
```http request
 curl -X POST 0.0.0.0:8080/api/upload -H "Content-Type: multipart/form-data" -F "file=@/Users/johnny-goat/temp/values.yml"
```

Response:
````http respone
{"etag":"d41ee712-4573-410c-bdc2-30bb80c6fd4b","file":"'/tmp/values.yml' uploaded!","size":4225,"status":"ok","time":"2024-10-01T21:42:39.303881-04:00"}
````


 
