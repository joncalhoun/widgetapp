# Web App to API Demo

This app is meant to help demonstrate how easy it can be to migrate a Go application that renders HTML into a JSON API, and how nearly all of the logic in the application will remain unchanged. It intentionally starts out with a pretty poor design and structure so that we can look at the benefits of each individual set of changes we will be making.

## Setup

To setup your local dev you will need to setup a PostgreSQL database. I provided a `setup.sql` file to help make that a little easier - you should be able to run it like this:

```
psql -f setup.sql
```

If you need help figuring our Postgres, I have a pretty in-depth series on using it here: <https://www.calhoun.io/using-postgresql-with-go/>
