# URL Scraper

Watch webpages for changes.

### Development

1. Create a Postgres database named `urlscraper`: `$ createdb urlscraper`.
1. Initialize and migrate the database: `go run db/main.go`.
1. Install the Node.js dependencies: `npm install`.
1. Rebuild the CSS: `npm run css`.
1. Rebuild the JS: `npm run js`.
1. Start the application: `go run main.go`.

### Deployment

The following environment variables must be set.

* `DSN` – The Postgres DSN.
* `CLIENT_ID` - The DO OAuth client ID.
* `CLIENT_SECRET` - The DO OAuth client secret.
* `REDIRECT_URL` - The DO OAuth redirect URL.
