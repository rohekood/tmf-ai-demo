# UI Architecture Document

## 1. Executive Summary
This document outlines the architecture for the `demo-ui` service, a modern web application designed to demonstrate the capabilities of the TMF API microservices. The architecture prioritizes type safety, maintainability, and clean separation of concerns using a Backend-for-Frontend (BFF) pattern.

## 2. Technology Stack

### Frontend
-   **Framework**: React 19 (via Vite)
-   **Language**: TypeScript (Strict Mode)
-   **Build Tool**: Vite
-   **Package Manager**: Yarn 4 (Berry)
-   **State Management**: React Hooks (Reducer, Context) + TanStack Query (v5)
-   **Data Grid**: TanStack Table (v8) + TanStack Virtual (for scrolling performance)
-   **Component Primitives**: Radix UI (Headless, Accessible)
-   **Icons**: Lucide React
-   **Styling**: Vanilla CSS (CSS Modules or utility classes if needed)
-   **Testing**: Vitest (Unit), React Testing Library (Component), Playwright (E2E)

### Backend-for-Frontend (BFF)
-   **Language**: Go (Golang) 1.24+
-   **Framework**: Standard Library (`net/http`) or lightweight router (e.g., `chi` or `gorilla/mux`)
-   **Role**: Proxy, Aggregation, Authentication Handler, Static File Server

### Infrastructure
-   **Containerization**: Docker (Multi-stage builds)
-   **Orchestration**: Docker Compose (Local Development)

### High-Level Architecture

```mermaid
graph TD
    User[User Browser] 
    User -->|HTTPS - Static Assets| UI_Server["UI Service (Nginx)"]
    User -->|HTTPS - API| BFF[Golang BFF Service]
    
    subgraph "Demo UI Layer"
        UI_Server -->|Serve| UI[React App Files]
    end
    
    subgraph "BFF Layer"
       BFF -->|RPC / Events| Rabbit[RabbitMQ]
    end
    
    subgraph "Internal Network"
        Rabbit <-->|Consumer/Publisher| Cust[Customer Management Service]
        Rabbit <-->|Consumer/Publisher| Party[Party Management Service]
    end
    
    BFF <-->|OIDC| Okta[Okta Identity Provider]
```

### 3.1 Backend-for-Frontend (BFF) Pattern
The BFF behaves as an API Gateway for the Frontend.
-   **Separation**: The UI and BFF are deployed separately.
    -   **Demo UI Container**: Nginx server hosting the compiled React application.
    -   **BFF Container**: Golang service handling API requests.
-   **Responsibilities**:
    -   **Auth**: BFF handles OAuth2/OIDC. UI redirects to BFF for login.
    -   **RabbitMQ RPC Client**: BFF converts HTTP requests from the UI into RabbitMQ messages.
    -   **CORS**: BFF must allow requests from the UI domain.

## 4. Key Design Decisions

### 4.1 State Management (No Redux/MobX)
-   **Server State (Data Sync)**: usage of **TanStack Query (React Query)**. This handles caching, deduplication, loading states, and refetching logic for all API data.
-   **Client State (UI)**: usage of **React Context + useReducer** for global UI state (e.g., theme, sidebar toggle, toast notifications). Local component state uses `useState`.

### 4.2 Authentication & Authorization
-   **Provider**: Okta.
-   **Flow**: Authorization Code Flow with PKCE.
-   **Implementation**:
    -   The BFF handles the OIDC exchange to keep tokens secure on the server side (Session Cookie pattern recommended for best security).
    -   Alternatively, if Client-Side Auth is preferred, the React app uses `@okta/okta-react`, but the BFF approach is safer.
    -   *Decision*: **BFF-handled Auth** (Session Cookie) is proposed for higher security, but Client-Side Auth is simpler for a "demo". Given the requirement for a Golang BFF, we will leverage it for security.

### 4.3 Testing Strategy
-   **Unit Tests**: Business logic functions (hooks, reducers, utils) tested with **Vitest**.
-   **Component Tests**: Rendered components tested with **React Testing Library**.
-   **Integration/E2E**: Critical user flows tested with **Playwright** against a running stack.
-   **Coverage**: 100% test coverage goal for business logic.

### 4.4 Stateless Architecture & Cluster Deployment

Both the **UI** and **BFF** components are designed to be **stateless** and **horizontally scalable**.

#### UI (React/Nginx)
-   **Stateless by Design**: The React app is a static build served by Nginx. No server-side state.
-   **CDN-Ready**: Static assets can be distributed via CDN for global availability.
-   **Cluster Deployment**: Multiple Nginx pods can run behind a load balancer with no session affinity required.

#### BFF (Golang)
-   **No Local State**: All state is externalized:
    -   **Session Data**: Stored in external Redis (not in-memory).
    -   **Correlation Data**: Stored in RabbitMQ reply queues (auto-deleted).
-   **Stateless Request Handling**: Each HTTP request is independent; any BFF instance can handle any request.
-   **Horizontal Scaling**:
    -   Deploy multiple BFF replicas behind a Kubernetes Service or Load Balancer.
    -   No session affinity (sticky sessions) required.
    -   Health checks via `/health` endpoint for liveness/readiness probes.

#### Deployment Requirements
| Component | Scaling Strategy | Session Affinity | State Storage |
|:----------|:-----------------|:-----------------|:--------------|
| UI (Nginx) | Horizontal (N replicas) | Not Required | None (static files) |
| BFF (Go) | Horizontal (N replicas) | Not Required | Redis (sessions), RabbitMQ (RPC) |

#### Kubernetes Considerations
-   **ReplicaSet/Deployment**: Both UI and BFF should be deployed as Kubernetes Deployments with `replicas >= 2` for high availability.
-   **Service Type**: Use `ClusterIP` for internal services, `LoadBalancer` or `Ingress` for external access.
-   **Pod Disruption Budget**: Configure PDBs to ensure availability during rolling updates.
-   **Resource Limits**: Define CPU/memory requests and limits for predictable scaling.

## 5. Folder Structure

```
services/
├── demo-ui/
│   ├── ui/                  # Frontend React App (formerly direct under demo-ui)
│   │   ├── src/
│   │   ├── Dockerfile
│   │   └── package.json
│   │
│   └── bff/                 # Golang API Gateway/BFF
│       ├── cmd/server/
│       ├── internal/
│       ├── go.mod
│       └── Dockerfile
```

## 6. Implementation Plan - Next Steps
1.  **Scaffold BFF**: Create `services/demo-bff` with Go.
2.  **Re-scaffold UI**: Re-initialize `services/demo-ui` using the TypeScript template.
3.  **Setup Dev Environment**: Configure `docker-compose` to run BFF + UI + Backends.
