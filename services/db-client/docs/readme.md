# Db-client

This service exposes http endpoints to insert/update/delete items in the database.

## Endpoints

### GET /health

Checks if database connection is still open.

**Input:** nil

**Output:**
`{"status": "ok"}`

### GET /venues/{id}

Joins location and venue data and returns relevant json

**Input:** UUID, ex: 81451352-c86f-4c91-94d3-2e2bb396a586

**Output:**

```json{
    "id": "81451352-c86f-4c91-94d3-2e2bb396a586",
    "name": "Foobar",
    "location": {
        "id": "c635983c-5595-4c10-be3e-dd3cf7d6ee40",
        "street": "Borgarfjordsgatan 99",
        "area": "Kista",
        "city": "Stockholm",
        "country": "Sweden",
        "zip": "164 25",
        "lat": 59.4067,
        "lng": 17.9452,
        "created_at": "2026-04-24T12:39:34.301862Z",
        "updated_at": "2026-04-24T12:39:34.301862Z",
        "deleted_at": null
    },
    "created_at": "2026-04-24T12:39:34.301862Z",
    "updated_at": "2026-04-24T12:39:34.301862Z"
}
```

### POST /venues/create

Expects a json payload in from the request, and inserts it into the database.

**Input:**

```json{
    "name": "Foobar",
    "street": "Borgarfjordsgatan 99",
    "area": "Kista",
    "city": "Stockholm",
    "country": "Sweden",
    "zip": "164 25",
    "lat": 59.4067,
    "lng": 17.9452
}
```

**Output**
Status 201 created, if input was valid.

## Build

Create the docker image with `docker build -t db-client .`

Run the container with `docker run --env-file .env -p 8081:8081 db-client`

Requires a .env file with SUPABASE_CONN_STRING.
