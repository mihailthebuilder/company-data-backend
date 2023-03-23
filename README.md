# Company Data Backend

Run the server locally:

```
go run main.go
```

Run Dockerised application locally, in production mode:
```
docker build -t company-data-backend-container .
docker run -p 8080:8080 --env-file .env --env GIN_MODE=release company-data-backend-container
```

Deploy to Hetzner server with Caprover:
```
caprover deploy
```

Test email endpoint
```
curl -X POST http://localhost:8080/email
```

Test sampler endpoint
```
curl -X POST http://localhost:8080/companies/sample -d '{"SicDescription":"Extraction of salt"}'
```