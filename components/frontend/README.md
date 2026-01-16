# IdeKube Frontend

A modern Vue 3 + Vuestic UI frontend for the IdeKube Kubernetes IDE management platform.

## Features

- 🎨 Built with Vuestic UI framework
- ⚡️ Vite for fast development and building
- 🔒 TypeScript support
- 📱 Responsive design
- 🔌 Axios for API communication
- 📦 Pinia for state management
- 🛣️ Vue Router for navigation

## Prerequisites

- Node.js >= 18.x
- Yarn package manager

## Development

### Install dependencies

```bash
make install
# or
yarn install
```

### Run development server

```bash
make dev
# or
yarn dev
```

The application will be available at `http://localhost:5173`

### Type checking

```bash
make type-check
# or
yarn type-check
```

### Linting

```bash
make lint
# or
yarn lint
```

## Building

### Build for production

```bash
make build
# or
yarn build
```

### Preview production build

```bash
make preview
# or
yarn preview
```

## Docker

### Build Docker image

```bash
make docker-build
```

This builds a production-optimized Docker image with nginx.

### Run in Docker (production)

```bash
make docker-run
```

The application will be available at `http://localhost:8080`

#### Configure Backend URL

The container accepts environment variables to configure the backend API endpoint:

```bash
docker run -d \
  -p 8080:80 \
  -e BACKEND_HOST=api.example.com \
  -e BACKEND_PORT=3000 \
  --name idekube-frontend \
  davidliyutong/idekube-frontend:latest
```

**Environment Variables:**
- `BACKEND_HOST` - Backend hostname or IP (default: `localhost`)
- `BACKEND_PORT` - Backend port (default: `3000`)
- `RESOLVER` - DNS resolver IP (default: `127.0.0.11` for Docker)

**Docker Compose Example:**

```yaml
version: '3.8'
services:
  frontend:
    image: davidliyutong/idekube-frontend:latest
    ports:
      - "80:80"
    environment:
      - BACKEND_HOST=backend
      - BACKEND_PORT=3000
    depends_on:
      - backend
  
  backend:
    image: your-backend-image
    ports:
      - "3000:3000"
```

**Kubernetes Example:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idekube-frontend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: idekube-frontend
  template:
    metadata:
      labels:
        app: idekube-frontend
    spec:
      containers:
      - name: frontend
        image: davidliyutong/idekube-frontend:latest
        ports:
        - containerPort: 80
        env:
        - name: BACKEND_HOST
          value: "idekube-backend-service"
        - name: BACKEND_PORT
          value: "3000"
```

### Run in Docker (development)

```bash
make docker-dev
```

This runs a development container with hot reload enabled.

### Stop Docker container

```bash
make docker-stop
```

## Project Structure

```
src/
├── assets/          # Static assets
├── components/      # Vue components
├── router/          # Vue Router configuration
├── stores/          # Pinia stores
├── utils/           # Utility functions
├── views/           # Page components
├── App.vue          # Root component
└── main.ts          # Application entry point
```

## Environment Variables

Create a `.env.local` file for local development:

```env
VITE_API_URL=http://localhost:3000
VITE_WS_URL=ws://localhost:3000
```

## Configuration

- `vite.config.ts` - Vite configuration
- `tsconfig.json` - TypeScript configuration
- `nginx.conf` - Nginx configuration for production
- `.eslintrc.cjs` - ESLint configuration

## API Integration

The frontend is configured to proxy API requests through Vite dev server or nginx in production:

- `/api` - REST API endpoints
- `/ws` - WebSocket connections

Configure backend URL in environment variables or `vite.config.ts`.

## License

MIT
