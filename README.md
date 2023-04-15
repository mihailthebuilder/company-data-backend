# Company Data Backend

Run the server locally:

```
go run .
```

Run Dockerised application locally, in production mode:

```
docker build -t company-data-backend-container .
docker run -p 8080:8080 --env-file .env company-data-backend-container
```

Deploy to Hetzner server with Caprover:

```
caprover deploy --default
```

Test endpoints...

- authenticate
```
curl -X POST http://localhost:8080/register -d '{"EmailAddress":"hello@test.com", "Request":"hello world"}'
```

- sample
```
curl -X POST http://localhost:8080/companies/sample -d '{"SicDescription":"Extraction of salt"}'
```
