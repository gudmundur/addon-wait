# Addon Wait

This is a utility that will wait on Heroku Postgres and Redis to become available after provisioning. The utility will attempt to connect either addon for up to five minutes before failing. To determine whether Heroku Postgres has become available, it attempts to connect to `DATABASE_URL`. Similarly for Heroku Redis, it attempts to connect to `REDIS_URL`.

For instructions on how to use with review apps, head over to [heroku/heroku-buildpack-addon-wait](https://github.com/heroku/heroku-buildpack-addon-wait).

## Usage

Run

```console
bin/addon-wait
```

The utility will exit with code `0` for success, `1` for failure.

