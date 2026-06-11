This repo contains the backend of BilligBar.
Each service in `./services` is ran as an independent docker container.

# Services

## Db-client

This service handles all communication with our remote Supabase DB.

Responsibilities:

- Provide endpoints for the frontend to filter on venues, display menus, open hours etc
- Handle everything related to Supabase Auth
- Provide endpoints for creating as well as reviewing submissions

Detailed docs: `./services/db-client/readme.md`

## Get-places-data

This service acts as a wrapper around the Google Places API and provides endpoints to fetch venue data such as coordinates, open hours, etc.

Detailed docs: `./services/get-places-data/readme.md`

## Image-to-json

This service provides a single endpoint which takes an image of a menu, passes it to Groq API, and returns a json representation of the menu.

Detailed docs: `./services/image-to-json/readme.md`
