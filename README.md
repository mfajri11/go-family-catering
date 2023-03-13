# GO family-catering

## Overview

This project is re-written of my solution for the individual final project for backend engineer track held by [yayasan anak bangsa bisa (YABB)](https://www.anakbangsabisa.org/). The problem was asking us to create a program that helps automating some business aspects of food ordering management, for example create a menu, make an order, add a new owner etc. Since I don't have any rights to expose the problem statement, I wouldn't expose any more details. Hopefully the code could help you understand the problem statements.

## Content

- [Features](#features)
- [Requirements](#requirements)
- [Architecture](#architecture)
- [Note](#note)
- [How to run](#how-to-run)

## Features

- [Open API Specification](https://swagger.io/specification/v2/) 2.0  `Swagger`
- Make command options and its documentation via `make` or `make help`
- CronJob
- Access Token and Refresh Token using [JWT](https://www.rfc-editor.org/rfc/rfc7519)
- Session id with random `string` and Redis
- CRUD operation (Postgres `raw sql`)
- [Easy configuration interface](./config/config.md) via yaml file
- Database migrations
- Request Validation
- middlewares [CORS](https://github.com/go-chi/cors) , [HTTP Rate Limiter](https://github.com/go-chi/httprate) (By IP and Session ID), custom logger, etc.
- Graceful shutdown
- more...

## Requirements

- [Golang](https://go.dev/)(1.7+ recommended)
- [Docker](https://www.docker.com)(Required if you use the container version)
- [Postgres](https://www.postgresql.org)(Automatically installed if you use container version)
- [Redis](https://redis.io)(Automatically installed if you use container version)
- [Mailhog](https://github.com/mailhog/MailHog)(Optinal you can use real smtp server, it will automatically installed if you use container version)

## Architecture

This project has 4 domain layers

- model
- repository
- service
- handler

## Note

#### HTTP API Documentation

Any api documentation can be found at `http://localhost:9000/swagger/index.html`

#### Mailer

if you won't use a fake smtp server like `mailhog` please change your host address of your chosen smtp server as shown at Listing.1 and delete line as shown as Listing.2, In case you are using real smtp server such as [gmail](https://gmail.com) and get `bad credentials` error while your credentials is actually correct, please activate [less secure apps](https://myaccount.google.com/lesssecureapps).

Listing.1

```yaml
# ./config/config.development.yaml
mailer:
  host: <localhost or you choosen smtp server ip address>
  port: <you choosen smtp server port number>
  # others lines goes here
```

if you already have `mailhog` and won't use the docker version of `mailhog` please change `host` to `localhost` and `port` to whatever you use, the default port is `1025` and delete mailhog service as show at Listing.2

Listing.2

```yaml
# ./docker-compose.yaml
mailhog:
  image: mailhog/mailhog
  container_name: mailhog
  ports:
   - 1025:1025
   - 8025:8025
  networks:
   - fcat_network
```

#### Postgres

if you already have postgres server running on your machine and want to use yours. please delete the postgres service at `./docker-compose.yaml` just like the instruction for deleting mailhog service but postgres instead and do not forget to setting up your host and port at `.config/config.development.yaml` according to your postgres.

#### Redis

for redis service it goes the same way as the [Postgres](#postgres) section does.

if you want to use your `redis.conf` please paste your content to `./config/redis.conf` or if you don't want to use password or `redis.conf` please delete line in Listing.3.

Listing.3

```yaml
# ./docker-compose.yaml
redis:
  # delete from this
  command: ["redis-server", "/etc/redis/redis.conf", "--save 30 1"]
  volumes:
    - ./config/redis.conf:/usr/local/etc/redis/redis.conf
  # to this
```

## How to run

###### clone

Download the repo and move to its directory in order to use `make` for `install`, `migrate` etc

```shell
# clonning the repo
$ git clone https://github.com/mfajri11/go-family-catering.git
# move to clonned repo directory
$ cd go-family-catering
```

###### Running

build all containers and attach to the go server only

```shell
# using a container
$ make install

# not using a container (make sure postgres, redis and mailhog is ready)
$ make install container=false # will only build go server
```

###### Destory

Stop and remove all built containers

```shell
# stop and remove all built containers by make install
$ make down

# to kill and non-container go server you can kill by its pid or via task manager (if you are using windows)
```

Applying down migrations

```bash
# will apply migration al the way down from active version of schema
$ make rollbacks
```
