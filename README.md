# BrainBlitz

## Requirements
[UseCase and Entity](document/requirement/requirements.md)

## [Graceful-Shutdown](https://github.com/nima-abdpoor/BrainBlitz/blob/develop/document/development/graceful-shutdown.md)

## How to Run
### step 1: Create shared network with docker:
```bash 
  docker network create bb-network
```
### step 2: Run each service:
```bash
  docker-compose -f deploy/user/development/docker-compose.yaml up -d
  docker-compose -f deploy/match/development/docker-compose.yaml up -d
  docker-compose -f deploy/game/development/docker-compose.yaml up -d
```

### step 3: Run shared services in the main directory:
```bash
  docker-compose up -d
```