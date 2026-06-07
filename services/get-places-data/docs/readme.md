# Get places data

This service acts as a proxy for the Google Places API. It exposes two endpoints. One is for searching for places using text like names and/or an address, which returns a list of all found places. The other endpoint takes an ID and returns detailed data about the place identified by that ID.

## Build

Create the docker image with `docker build -t get-places-data .`

Run the container with `docker run --env-file .env -p 8082:8082 get-places-data`

Requires a .env file with GOOGLE_API_KEY and Google places api (New) activated in Google Cloud.

## Endpoints

### GET /findplace?name={name}

Finds places matching the searchterm and returns their ids.

#### Example Call:

```console
curl "http://localhost:8082/findplace?name=Lion+Bar"
```

#### Example Response:

```json
[
  {
    "id": "ChIJR0uPnGidX0YRT7RD-y_cayI",
    "name": "Lion Bar",
    "address": "Sveavägen 74, 113 59 Stockholm, Sweden"
  },
  {
    "id": "ChIJxVwLaE2dX0YRP6HN55Hp8OE",
    "name": "Lion bar",
    "address": "Tulegatan 7, 172 78 Sundbyberg, Sweden"
  }
]
```

### GET /placeinfo?id={id}

Retrieves data from Google Places API for the specified place.

#### Example Call:

```console
curl "http://localhost:8082/placeinfo?id=ChIJR0uPnGidX0YRT7RD-y_cayI"
```

#### Example Response

```json
{
  "place_id": "ChIJR0uPnGidX0YRT7RD-y_cayI",
  "name": "The Mocked Pub",
  "street": "Testgatan 1",
  "city": "Stockholm",
  "rating": 4.8,
  "opening_hours": [
    {
      "day": 1,
      "open": "11:00",
      "close": "23:00"
    },
    {
      "day": 2,
      "open": "11:00",
      "close": "23:00"
    },
    {
      "day": 5,
      "open": "11:00",
      "close": "02:00"
    }
  ]
}
```
