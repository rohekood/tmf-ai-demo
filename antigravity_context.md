# Antigravity Context

## Project Overview
This is a monorepo for a **TMForum-inspired catalog-driven ordering system**. 
The system allows clients to order different services (e.g., broadband internet). 
**Scope**: covers processes up to order creation. Order handling itself is currently out of scope.

## Technology Stack
- **Languages**: Go
- **Database**: PostgreSQL
- **ORM**: GORM
- **Migrations**: golang-migrate
- **Infrastructure**: Kubernetes, Microservices
- **Communication**: Asynchronous APIs

## Architecture
- **Microservices**: The application is built as a set of microservices.
- **TMF Alignment**: Follows TMF application structure.
- **Asynchronous**: All APIs are designed to be asynchronous.

## Microservices Architecture
The system is composed of the following TMF-aligned microservices:

### 1. Product Catalog Management (TMF620)
- **Role**: Manages the lifecycle of product specifications and offerings.
- **Key Function**: defining what can be sold to customers.

### 2. Service Catalog Management (TMF633)
- **Role**: Manages the lifecycle of service specifications.
- **Key Function**: Defining the technical services that back the commercial products (e.g., broadband speed profiles).

### 3. Service Qualification Management (TMF645)
- **Role**: Checks if a service can be provided at a specific location or for a specific customer.
- **Key Function**: Broadband availability check.

### 4. Geographic Address Management (TMF673)
- **Role**: Manages and validates addresses.
- **Key Function**: Standardizing addresses for service qualification and delivery.

### 5. Party Management (TMF632)
- **Role**: Manages information about individuals and organizations.
- **Key Function**: Centralized identity and profile management for all actors.

### 6. Customer Management (TMF629)
- **Role**: Manages customer specific information and interactions.
- **Key Function**: Handling customer accounts and status.

### 7. Product Ordering Management (TMF622)
- **Role**: Manages the order lifecycle from creation to completion.
- **Key Function**: Capturing and processing customer orders (in-scope: until creation).

## Directory Structure
- `/`: Root directory

## Key Workflows
[TBD: Essential development commands and workflows]

## Development Standards
- **Testing**: All functionality must be covered with unit tests.
