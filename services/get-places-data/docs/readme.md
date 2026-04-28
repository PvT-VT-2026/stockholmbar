# Get places data

This service acts as a proxy for the Google Places API. It exposes two endpoints. One is for searching for bars using text like names or/and a address, which returns a list of all found ids. The other endpoint takes a id and returns data about the place identified by the id.

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
  "name": "Lion Bar",
  "street": "Sveavägen  74",
  "area": "",
  "city": "Stockholm",
  "country": "Sweden",
  "zip": "113 59",
  "lat": 59.3398946,
  "lng": 18.0597901,
  "rating": 3.6,
  "opening_hours": [
    "Monday: 1:00 PM – 3:00 AM",
    "Tuesday: 1:00 PM – 3:00 AM",
    "Wednesday: 1:00 PM – 3:00 AM",
    "Thursday: 1:00 PM – 3:00 AM",
    "Friday: 1:00 PM – 3:00 AM",
    "Saturday: 1:00 PM – 3:00 AM",
    "Sunday: 1:00 PM – 3:00 AM"
  ]
}
```
