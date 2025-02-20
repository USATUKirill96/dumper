# Dumper

A utility for creating and loading PostgreSQL database dumps.

## Features

- Create database dumps from different environments
- Automatic management of local PostgreSQL in Docker
- Load dumps into local database
- Support for multiple environments
- Debug mode for detailed information

## Installation

1. Clone the repository:
```bash
git clone https://github.com/your-username/dumper.git
cd dumper
```

2. Install dependencies:
```bash
go mod download
```

3. Copy example configuration and edit it:
```bash
cp config.example.yaml config.yaml
```

## Configuration

Edit `config.yaml` and specify your database connection parameters:

```yaml
environments:
  - name: dev
    db_dsn: postgres://user:password@dev-host:5432/database?sslmode=disable
  
  - name: stage
    db_dsn: postgres://user:password@stage-host:5432/database?sslmode=verify-full
```

## Usage

1. Run the utility:
```bash
go run . [--debug]
```

2. Select environment from the list
3. Use available commands:
   - Create dump
   - Load dump into local database
   - Change environment

## Requirements

- Go 1.21 or higher
- Docker
- PostgreSQL client (for creating dumps) 