# g37-lanches

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/golang.org/x/example)

Serviço de controle de pedidos desenvolvido em Golang.

## Tech Stack

**API:** Go

**Infra:** PostgreSQL, SQS

<br/>

## Requisitos

- go 1.20
- docker
- kubernetes cluster (Docker desktop)
- kubectl

<br/>

## Execução com Docker Compose

Build image

```bash
  docker build -t lanches-api:latest .
```

Subir dependências
```bash
  docker-compose up -d
```
<br/>

## Execução com Kubernetes

Entrar na pasta do Kubernetes
```bash
  cd k8s
```

Criar Persistent Volume
```bash
  kubectl apply -f pv-volume.yaml
```

Criar Persistent Volume Claim
```bash
  kubectl apply -f pv-claim.yaml
```

Criar Postgres Config Map
```bash
  kubectl apply -f postgres-config-map.yaml
```

Criar Postgres Service
```bash
  kubectl apply -f postgres-service.yaml
```

Criar Postgres Deployment
```bash
  kubectl apply -f postgres-deployment.yaml
```

Criar API Service
```bash
  kubectl apply -f api-service.yaml
```

Criar API Deployment
```bash
  kubectl apply -f api-deployment.yaml
```

## Documentação
[Documentation](https://github.com/IgorRamosBR/g37-techchallenge/tree/master/docs)


## Arquitetura
Clean Architecture com a estrutura de pastas baseada no [Standard Go Project Layout](https://github.com/golang-standards/project-layout#go-directories) 

```bash
├── cmd
├── configs
├── docs
├── internal
|   |── api
|   |── controllers
|   │   ├── _api
|   │   ├── application
|   ├── core
|   │   ├── entities
|   │   ├── usecases
|   ├── infra
|   │   ├── drivers
|   │   ├── gateways
├── k8s
├── migrations
```

## Arquitetura
![cloud-architecture](https://github.com/nobrucarneiro/techchallenge/assets/46383235/d8f89139-418b-4442-9e36-f8ee7a6c34aa)

