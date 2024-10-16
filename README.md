# DBA Self Service
## Tech Stack

- **Frontend:** SvelteKit
- **Backend:** Go
- **Database:** PostgreSQL (for user information and logging) and MS SQL (for updations and validation)
- **Containerization:** Docker

## Setup

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.22 or later)
- [Node.js](https://nodejs.org/) (version 20 or later)
- [Docker](https://www.docker.com/) (for containerized deployment)
- [PostgreSQL](https://www.postgresql.org/download/) (for recording request logs)


### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/samarth-sh/DBA-SelfService.git
   ```
### Running the Application

#### Using Docker

1. **Build and run**:

    ```bash
    docker-compose build --no-cache
    docker-compose up
    ```

2. **Access the application**:

    - Frontend: `http://localhost:5173`
    - Backend: `http://localhost:8080`

#### Without Docker

1. **Backend**:

    - Navigate to the `backend` directory:

      ```bash
      cd go-backend/
      ```

    - Install dependencies and start the server:

      ```bash
      go mod tidy
      go run main.go
      ```

    - The backend should be running at `http://localhost:8080`.

2. **Frontend**:

    - Navigate to the `sveltekit-frontend` directory:

      ```bash
      cd sveltekit-frontend/
      ```

    - Install dependencies and start the development server:

      ```bash
      npm install
      npm run build
      npm run dev
      ```

    - The frontend should be running at `http://localhost:5173`.

### For monitoring

Use [DBeaver](https://dbeaver.com/download/) or [Azure Data Studio](https://learn.microsoft.com/en-us/azure-data-studio/download-azure-data-studio?view=sql-server-ver16&tabs=win-install%2Cwin-user-install%2Credhat-install%2Cwindows-uninstall%2Credhat-uninstall) to view and monitor the databases

## Acknowledgments

- [SvelteKit](https://kit.svelte.dev/) for the frontend framework.
- [Go](https://golang.org/) for the backend framework.
- [Docker](https://www.docker.com/) for containerized deployment.
- [PostgreSQL](https://www.postgresql.org/) and [MS-SQL](https://www.microsoft.com/en-in/sql-server/sql-server-downloads) for database setup and configurations.
