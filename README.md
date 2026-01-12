# Nyx Backend

## Description
Nyx is a backend service designed to manage items, claims, and user interactions in a streamlined manner. It utilizes PostgreSQL for data storage and provides a RESTful API for client interactions.

## Tech Stack
- **Programming Language:** Go
- **Framework:** Gin
- **Database:** PostgreSQL
- **Containerization:** Docker
- **Task Management:** Taskfile

## Setup
1. **Clone the Repository**
   ```sh
   git clone https://github.com/KiranRajeev-KV/nyx-backend.git
   cd nyx-backend
   ```

2. **Setup the project**
   Ensure you have Go, Docker, and [Taskfile](https://taskfile.dev/docs/installation) installed on your machine.
   ```sh
   task setup
   ```
   This will install air (for hot reload), goose (for migrations), lefthook (for git hooks), and sqlc (for code generation).

3. **Run Migrations**
   Apply the database migrations:
   ```sh
   task up
   ```

## Development
To start the development server with hot reload, run:
```sh
task dev
```
This will start the server and allow you to make changes without restarting it manually.


## Authors
- Kiran Rajeev K V - [GitHub](www.github.com/KiranRajeev-KV)



