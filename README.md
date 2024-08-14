# Password Reset Portal
## Tech Stack

- **Frontend:** SvelteKit
- **Backend:** Go
- **Database:** SQLite (for user information and logging)
- **Containerization:** Docker

## Setup

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.23 or later)
- [Node.js](https://nodejs.org/) (version 14 or later)
- [Docker](https://www.docker.com/) (optional, for containerized deployment)
- [SQLite](https://www.sqlite.org/download.html) (or any compatible SQL database)


### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/samarth-sh/DBA-SelfService.git
   cd password-reset-portal
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
      npm run dev
      ```

    - The frontend should be running at `http://localhost:5173`.

## Logging and Monitoring

All password update requests and actions are logged for auditing and monitoring purposes. The logs are stored in a dedicated table within the database with the following fields:

- `request_id`: A unique identifier for each request.
- `request_type`: Defaults to 'update password'.
- `request_status`: Indicates the outcome of the request (e.g., success, failure).
- `message`: A brief description of the action performed.
- `timestamp`: The date and time when the action occurred.

## Acknowledgments

- [SvelteKit](https://kit.svelte.dev/) for the frontend framework.
- [Go](https://golang.org/) for the backend framework.
- [Docker](https://www.docker.com/) for containerized deployment.
