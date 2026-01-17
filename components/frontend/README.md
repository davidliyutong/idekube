# IDEKube Frontend

A modern Vue 3 + Vuestic UI frontend for the IDEKube Kubernetes IDE management platform.

## Features

- 🎨 Built with Vuestic UI framework
- ⚡️ Vite for fast development and building
- 🔒 TypeScript support with strict mode
- 📱 Responsive design
- 🔌 Axios for API communication with auto-retry
- 📦 Pinia for state management
- 🛣️ Vue Router 4 with navigation guards
- 🔐 JWT authentication with auto-refresh
- 🔄 OpenAPI code generation from swagger.json
- 🐳 Docker support with nginx

## Prerequisites

- Node.js >= 18.x
- Yarn >= 1.22.0
- Docker (optional, for containerized deployment)

## Project Structure

```
src/
├── api/              # API client and generated code
├── assets/           # Static assets
├── components/       # Reusable components
├── layouts/          # Layout components (Auth, App)
├── pages/            # Page components
│   ├── auth/         # Login, Register, Forgot Password
│   ├── dashboard/    # Dashboard
│   ├── workspaces/   # Workspace management
│   ├── templates/    # Template management
│   ├── users/        # User management (admin)
│   ├── organizations/# Organization management
│   ├── volumes/      # Volume management
│   ├── settings/     # Settings (admin)
│   ├── webhooks/     # Webhooks (admin)
│   ├── api-keys/     # API key management
│   └── profile/      # User profile
├── router/           # Vue Router configuration
├── stores/           # Pinia stores
├── styles/           # Global styles
├── types/            # TypeScript types
└── utils/            # Utility functions
```

## Development

### Install dependencies

```bash
make install
# or
yarn install
```

### Generate API Client

Before running the development server, generate the API client from the controller's swagger.json:

```bash
make generate-api
# or
./scripts/generate-api.sh
```

This will create TypeScript client code in `src/api/client/` based on the OpenAPI specification from `../controller/docs/api/swagger.json`.

**Requirements:**
- Java 8+ installed (required by OpenAPI Generator)
- Controller's swagger.json must exist at `../controller/docs/api/swagger.json`

**Note for macOS users:** If you encounter Java path issues with jenv, the script will automatically use the system Java installation via `/usr/libexec/java_home`.

### Run development server

```bash
make dev
# or
yarn dev
```

The application will be available at `http://localhost:5173`

The dev server includes:
- Hot Module Replacement (HMR)
- API proxy to backend (configurable via VITE_API_BASE_URL)
- WebSocket proxy for real-time features

### Environment Variables

Copy `.env.example` to `.env.development` and configure:

```env
VITE_API_BASE_URL=http://localhost:3000
VITE_APP_TITLE=IDEKube
VITE_ENABLE_MOCK=false
```

### Type checking

```bash
make type-check
# or
yarn type-check
```

### Linting & Formatting

```bash
make lint      # Run ESLint
make format    # Run Prettier
# or
yarn lint
yarn format
```

## Building

### Build for production

```bash
make build
# or
yarn build
```

The build output will be in the `dist/` directory.

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

This builds a production-optimized Docker image with:
- Multi-stage build (Node.js builder + nginx runtime)
- Optimized bundle size with tree-shaking
- Gzip compression
- Environment variable substitution for backend URL

### Run in Docker (production)

```bash
make docker-run
```

The application will be available at `http://localhost:8080`

### Configure Backend URL

The container accepts environment variables to configure the backend API endpoint:

```bash
docker run -d \
  -p 8080:80 \
  -e BACKEND_HOST=api.example.com \
  -e BACKEND_PORT=3000 \
  -e RESOLVER=8.8.8.8 \
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

The frontend uses a type-safe API client auto-generated from the controller's OpenAPI specification.

### Generate API Client

```bash
make generate-api
```

This creates TypeScript client code in `src/api/client/` with:
- Type-safe API endpoint methods
- Request/response type definitions
- Automatic JWT token handling
- Built-in error handling

### Using the API Client

```typescript
import { DefaultApi, apiClient } from '@/api';
import type { ModelsWorkspace } from '@/api';

const api = new DefaultApi(undefined, '', apiClient);
const response = await api.workspacesGet();
const workspaces: ModelsWorkspace[] = response.data.workspaces || [];
```

**See [docs/API_CLIENT.md](docs/API_CLIENT.md) for comprehensive usage guide.**

### API Proxy Configuration

The frontend proxies API requests through Vite dev server or nginx in production:

- `/api` - REST API endpoints
- `/ws` - WebSocket connections

Configure backend URL in environment variables or `vite.config.ts`.

## License

MIT
