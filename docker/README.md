# Dockerized Indigo Server Stack

This directory contains a Docker Compose configuration for running the complete Indigo stack (API gateway, PostgreSQL, Valkey/Redis, and the admin frontend).

## Quick Start

1. Copy the example environment file and adjust values as needed:

   ```bash
   cp docker/.env.example docker/.env
   ```

   The `.env` file holds database credentials, secrets, and frontend configuration. Never commit the populated `.env` file.

2. Build and start the stack:

   ```bash
   docker compose --project-directory "$(pwd)/.." -f docker/docker-compose.yml up --build
   ```

3. Access the services:

   - API Gateway: <http://localhost:8080>
   - Admin Frontend: <http://localhost:3000>

4. Shut everything down when finished:

   ```bash
   docker compose --project-directory "$(pwd)/.." -f docker/docker-compose.yml down
   ```

The compose file references:

- `apps/indigo-api-gateway/Dockerfile` for the Go API service
- `indigo-frontend/Dockerfile` for the Next.js admin UI
- Standard images for PostgreSQL and Valkey (Redis-compatible)
