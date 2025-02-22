# Smrv2 API

Smrv2 is a web scraper that scrapes schedules from the official website. The API is built with Go and Fiber.

[![Go](https://img.shields.io/github/go-mod/go-version/savioruz/smrv2-api)](https://golang.org/)
![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/savioruz/smrv2-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/savioruz/smrv2-api)](https://goreportcard.com/report/github.com/savioruz/smrv2-api)
[![GitHub issues](https://img.shields.io/github/issues/savioruz/smrv2-api)](https://goreportcard.com/report/github.com/savioruz/smrv2-api)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/savioruz/smrv2-api)](https://goreportcard.com/report/github.com/savioruz/smrv2-api)

## Table of Contents

- [Features](#features)
- [Deployment](#deployment)
- [Requirements](#requirements)
- [Installation](#installation)
- [Usage](#usage)
  - [Running the API](#running-the-api)
  - [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgements](#acknowledgements)

## Features

- Scrapes schedules from the official website
- Caches data in Redis
- Cron job to update data
- API Documentation with Swagger
- Docker support

## Deployment

- ### Koyeb
[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?name=smrv2-api&repository=savioruz%2Fsmrv2-api&branch=main&builder=dockerfile&instance_type=free&regions=was&env%5BAPP_DOMAIN%5D=&env%5BAPP_ENV%5D=production&env%5BAPP_KEY_PASSWORD%5D=%7B%7B+secret.APP_KEY_PASSWORD+%7D%7D&env%5BAPP_LOG_LEVEL%5D=1&env%5BAPP_NAME%5D=smrv2-api&env%5BAPP_PORT%5D=3000&env%5BAPP_SALT_PASSWORD%5D=%7B%7B+secret.APP_SALT_PASSWORD+%7D%7D&env%5BDB_HOST%5D=&env%5BDB_NAME%5D=&env%5BDB_PASSWORD%5D=%7B%7B+secret.DB_PASSWORD+%7D%7D&env%5BDB_PORT%5D=5432&env%5BDB_SSL_MODE%5D=require&env%5BDB_TIMEZONE%5D=Asia%2FJakarta&env%5BDB_USER%5D=&env%5BJWT_ACCESS_EXPIRY%5D=1d&env%5BJWT_REFRESH_EXPIRY%5D=30d&env%5BJWT_SECRET%5D=%7B%7B+secret.JWT_SECRET+%7D%7D&env%5BRABBITMQ_HOST%5D=&env%5BRABBITMQ_PASSWORD%5D=%7B%7B+secret.RABBITMQ_PASSWORD+%7D%7D&env%5BRABBITMQ_PORT%5D=5672&env%5BRABBITMQ_USERNAME%5D=&env%5BRABBITMQ_VHOST%5D=&env%5BREDIS_DB%5D=0&env%5BREDIS_HOST%5D=&env%5BREDIS_PASSWORD%5D=%7B%7B+secret.REDIS_PASSWORD+%7D%7D&env%5BREDIS_PORT%5D=18643&env%5BSMTP_HOST%5D=smtp.gmail.com&env%5BSMTP_PASSWORD%5D=%7B%7B+secret.SMTP_PASSWORD+%7D%7D&env%5BSMTP_PORT%5D=465&env%5BSMTP_USERNAME%5D=&ports=3000%3Bhttp%3B%2F&hc_protocol%5B3000%5D=tcp&hc_grace_period%5B3000%5D=5&hc_interval%5B3000%5D=30&hc_restart_limit%5B3000%5D=3&hc_timeout%5B3000%5D=5&hc_path%5B3000%5D=%2F&hc_method%5B3000%5D=get)

## Requirements

- Go 1.23+
- PostgreSQL
- RabbitMQ
- Redis
- Docker
- Docker Compose
- Make
- [migrate](https://github.com/golang-migrate/migrate-cli) for database migrations
- [air](https://github.com/air-verse/air) for hot reloading while developing

## Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/savioruz/smrv2-api.git
    cd smrv2-api
    ```

2. **Environment Variables:**

   Create a `.env` file in the root directory and add the following:

    ```bash
    cp .env.example .env
    ```

## Usage

### Running Application

You can run the API using Docker or directly with Make.

### Docker

1. **Run the application:**

    ```bash
    make docker.run
    ```

2. **Stop the application:**

    ```bash
    make docker.stop
    ```

You need to have PostgreSQL, RabbitMQ, and Redis running on your machine.

### Docker Compose (Recommended)

1. **Build the application:**

    ```bash
    make dc.build
    ```

2. **Run the application:**

    ```bash
    make dc.up
    ```

3. **Stop the application:**

    ```bash
    make dc.down
    ```

For production, don't forget to set the environment in `docker-compose.yml` if you want to expose the application to the public.

### API Documentation

Swagger documentation is available at: http://localhost:3000/swagger.

![Preview](/assets/preview.png)

## Project Structure

```
.
├── assets/                # Asset files and images
├── cmd/                   # Application entry points
├── db/                    # Database related files
├── docs/                  # Project documentation
├── internal/              # Private application code
│   ├── builder/           # Builder patterns
│   ├── dao/               # Data Access Objects
│   │   ├── entity/        # Entity models
│   │   └── model/         # Model for request and response
│   ├── delivery/          # HTTP handlers and routes
│   ├── gateway/           # External service integrations
│   ├── repository/        # Data storage implementations
│   └── service/           # Business logic layer
├── pkg/                   # Public shared packages
│   ├── config/            # Configuration management
│   └── helper/            # Helper utilities
└── test/                  # Test files
```

## Contributing

Feel free to open issues or submit pull requests with improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Fiber](https://github.com/gofiber/fiber)
