# PROMPTS

## 1

help me create a helm application that provides managed cloud ide, like the gitpod thing:

- the application comes with some bundled components:
  - a postgresql server to store user credentials, objects, etc.
  - a rabbitmq message queue for inter service communication
- the application will also have these services:
  - a controller service that handles crd operations and other core componets
  - a rbac microservice that handles permission
  - a housekeeping service that manages lifetime of containers
  - a/multiple frontend that hosts the ui

currently, the project code name is `idekube`, you can useit for the names

avoid creating useless documents

for now, only creaet necessary helm template files and values.yaml, dont do the testing, but write down instructions about how to test

## 1.1

do some modification to current helm app:

- make sure the service account has namespace admin and is passed to idekube-housekeeper
- it is `housekeeper`, not `housekeeping`
- the images of controller, rbac, housekeeper, frontend are named as davidliyutong/idekube-{{role}}, and should be variable
- make sure certificate issuer a variable
- make sure ingress class a variable
- make sure ingress hostname a variable

### 1.2

do not use helm dependencies, instead create template to install the two services

## 1.3

translate the docs to english

## 2

- create manifests/third_party/objectstorage/* to run a [RustFS](https://rustfs.com/) server as a statefulset with a service in kubernetes, with persistent volume claims for storage
- create manifests/third_party/registry/* to run a [angos](https://github.com/project-angos/angos) in kubernetes, with persistent volume claims for storage

## 2.1

for longhorn, traefik, cert-manager, generate install.sh for each

for cert-manager, generate cloudflare cluster-issuer and cloudflare-api-secret.yaml example

## 3

now create a skeleton go project in a sub-directory `idekube-controller`, it will be a controller program that deals with crd operations:

- the controller is cloud-native, configured via env variables
- the controller is written in golang
- the controller image is built by Dockerfile
- the controleller access k8s api, uses service account
- the controller access to postgresql and rabbitmq

several other components, idekube-{housekeeper}/{rbac} also uses the same skeleton, so make it modular and duplicate it for these components

finally, create a blank vue project frontend at idekube-frontend 

### 3.1

now create a skeleton vuestic project in a sub-directory `components/frontend`, it will be a frontend container that runs the UI:

the reference nginx.conf is:

```
server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (optional - configure backend URL)
    location /api {
        proxy_pass http://backend:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket proxy (optional)
    location /ws {
        proxy_pass http://backend:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 10240;
    gzip_proxied expired no-cache no-store private auth;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json;
    gzip_disable "MSIE [1-6]\.";
}
```

reference Makefile:

```Makefile
..PHONY: install dev build preview lint clean docker-build docker-run

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "latest")

# Install dependencies
install:
	npm install

# Run development server
dev:
	npm run dev

# Build for production
build:
	npm run build

# Preview production build
preview:
	npm run preview

# Lint code
lint:
	npm run lint

# Type check
type-check:
	npm run type-check

# Clean build artifacts
clean:
	rm -rf dist node_modules

# Build Docker image
docker-build:
	docker build -t davidliyutong/idekube-frontend:${VERSION} .
	docker tag davidliyutong/idekube-frontend:${VERSION} davidliyutong/idekube-frontend:latest

# Run Docker container
docker-run: docker-build
	docker run -p 8080:80 davidliyutong/idekube-frontend:latest
```

requirements:

- use the yarn
- use veustic ui framework
- generate Dockerfile to build the image