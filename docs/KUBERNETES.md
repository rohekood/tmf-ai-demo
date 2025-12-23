# Kubernetes Deployment Guide

This guide explains how to configure `customer-management` and `party-management` services with Postgres and RabbitMQ connections in Kubernetes.

## 1. Create a Secret for Sensitive Data
Avoid using plain `ConfigMaps` for passwords. Use a `Secret`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tmf-secrets
type: Opaque
stringData:
  # Connection Strings (format: postgres://user:password@host:port/db)
  CUSTOMER_DB_URL: "postgres://user:password@postgres-host:5432/customer_db?sslmode=disable"
  PARTY_DB_URL: "postgres://user:password@postgres-host:5432/party_db?sslmode=disable"
  
  # RabbitMQ (format: amqp://user:password@host:port/)
  RABBITMQ_URL: "amqp://user:password@rabbitmq-host:5672/"
```

## 2. Deploy Customer Management
The Customer Management service expects `DB_URL` and `RABBIT_URL`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: customer-management
spec:
  replicas: 1
  selector:
    matchLabels:
      app: customer-management
  template:
    metadata:
      labels:
        app: customer-management
    spec:
      containers:
        - name: server
          image: ghcr.io/<your-org>/customer-management:latest
          env:
            - name: DB_URL
              valueFrom:
                secretKeyRef:
                  name: tmf-secrets
                  key: CUSTOMER_DB_URL
            - name: RABBIT_URL
              valueFrom:
                secretKeyRef:
                  name: tmf-secrets
                  key: RABBITMQ_URL
          ports:
            - containerPort: 8080
```

## 3. Deploy Party Management
The Party Management service expects `POSTGRES_URL` and `RABBITMQ_URL`.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: party-management
spec:
  replicas: 1
  selector:
    matchLabels:
      app: party-management
  template:
    metadata:
      labels:
        app: party-management
    spec:
      containers:
        - name: server
          image: ghcr.io/<your-org>/party-management:latest
          env:
            - name: POSTGRES_URL
              valueFrom:
                secretKeyRef:
                  name: tmf-secrets
                  key: PARTY_DB_URL
            - name: RABBITMQ_URL
              valueFrom:
                secretKeyRef:
                  name: tmf-secrets
                  key: RABBITMQ_URL
          ports:
            - containerPort: 8080
```

> [!NOTE]
> Ensure that the variable names (`DB_URL`, `POSTGRES_URL`) match exactly what is defined in the application code.
