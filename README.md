# Company Data Backend

Run the server locally:

```
make run
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

- sample

```
curl -X POST http://localhost:8080/v2/companies/sample -d '{"SicDescription":"Activities of mortgage finance companies"}'
```
