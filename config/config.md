# Config and Environment Variables

## Config variables

The Listing.2 show all possible configurable value that can be written through `./config/config.development.yaml`  also  please notice in this markdown I use format `<outerKey>.<key>` for easier identification while they are actually written as shown at Listing.1.

Listing.1

```yaml
<outerKey>:
	<key>
```

| key                                  | type   | status   | example                             | default                             |
| ------------------------------------ | ------ | -------- | ----------------------------------- | ----------------------------------- |
| app.name                             | string | required | family-catering                     | -                                   |
| app.version                          | float  | required | 1.0                                 | -                                   |
| web.pagination-limit                 | int    | optional | 25                                  | 10                                  |
| web.allowed-origins                  | array  | optional | [ui.family-catering.com]            | [http://\*,https://\*]              |
| web.allowed-methods                  | array  | optional | [GET,POST]                          | [GET,POST,PUT,DELETE,OPTIONS]       |
| web.allowed-headers                  | array  | optional | [Authorization]                     | [Accept,Auhtorization,Content-Type] |
| web.max-age                          | int    | optional | 100                                 | 300                                 |
| web.limit-general-request-per-minute | int    | optional | 100                                 | 100                                 |
| web.access-token-ttl                 | string | optional | 5m                                  | 15m                                 |
| web.refresh-token-tll                | string | optional | 2160h (90 days)                     | 2160h (90 days)                     |
| server.host                          | string | required | localhost                           | -                                   |
| server.port                          | string | optional | 9000                                | 9000                                |
| server.read-timeout                  | string | optional | 20s                                 | 10s                                 |
| server.write-timeout                 | string | optional | 30s                                 | 10s                                 |
| server.shutdown-timeout              | string | optional | 5s                                  | 3s                                  |
| log.level                            | string | optional | debug                               | info                                |
| postgres.host                        | string | required | localhost                           | -                                   |
| postgres.port                        | string | optional | 5432                                | 5432                                |
| postgres.max-open-connection         | int    | optional | 25                                  | 10                                  |
| postgres.max-idle                    | int    | optional | 10                                  | 5                                   |
| postgres.max-lifetime                | string | optional | 1h                                  | 0s                                  |
| redis.host                           | string | required | localhost                           | -                                   |
| redis.port                           | int    | optional | 6379                                | 6379                                |
| redis.database-name                  | int    | optional | 2                                   | 0                                   |
| redis.max-retries                    | int    | optional | 10                                  | 3                                   |
| redis.poolsize                       | int    | optional | 15                                  | 10                                  |
| mailer.host                          | string | required | localhost                           | -                                   |
| mailer.port                          | int    | required | 1025                                | -                                   |
| mailer.support-email                 | string | required | support.family-catering@example.com | -                                   |
| mailer.template-forgot-password      | string | optional | your_forgot_password_template.txt   | forgot_password_template.txt        |

if you are using the config for `staging` or `production` environment you can copy the `config.development.yaml` to `config.staging.yaml` or `config.producion.yaml` and setting up your configurable value based on its environment and also please set the `FCAT_ENV` to `staging` or `production` which will be explain at section [Environment variable](#environment-variable)

## Environment variable

Please do not hardcode your secret keys or include it at your repository.  The `./.env` is only used for simplicity and this repo does not support reading secret from secrets management apps.

if you use `go` app without docker-container, please set the environment variables because they are required. you can set the environment variables by preceding they `key value pair` with `go` command as show at Listing.2

Listing.2

```shell
$ FCAT_ENV=development PG_USER=root ... go run ./cmd/main.go
```

| key                      | type   | status   | example                         |
| ------------------------ | ------ | -------- | ------------------------------- |
| FCAT_ENV                 | string | required | development                     |
| PG_USER                  | string | required | postgres-username               |
| PG_PASSWORD              | string | requried | postgres-secret                 |
| PG_DATABASE              | string | required | database-name                   |
| REDIS_PASSWORD           | string | required | redis-secret                    |
| MAILER_EMAIL             | string | requried | cs.family-catering@example.com  |
| MAILER_PASSWORD          | string | required | cs_family_catering_email_secret |
| SECRET_KEY_ACCESS_TOKEN  | string | required | secret-key-access-token         |
| SECRET_KEY_REFRESH_TOKEN | string | required | secret-key-refresh-token        |
