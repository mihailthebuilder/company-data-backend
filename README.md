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

Send request:
```
http://localhost:8080/companies/sample?SicDescription=Cultural%20education
```