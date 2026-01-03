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

### 1.1

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