# Project Structure Documentation

## Overview

A Go microservice template built on `txix-open/isp-kit`, implementing **Clean Architecture**. The framework provides bootstrap, cluster communication, database access (`dbrx`, `dbx`), gRPC/HTTP servers, RabbitMQ integration (`grmqx`), observability.

---

## Package Responsibilities

### Root Module (`main.go`)

**Location:** `main.go`

**Responsibility:** Entry point. Bootstraps the service via `bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors(), cluster.GrpcTransport)`, instantiates `assembly.Assembly`, registers runners/closers, sets up shutdown handler, and calls `app.Run()`. Intentionally thin — delegates all initialization to `assembly`.

---

### `assembly` Package

**Location:** `assembly/`

**Responsibility:** Composition root. Wires infrastructure components and dependencies between `isp-kit/bootstrap` and the application layers (`repository`, `service`, `controller`).

#### `assembly.go`
- **Infrastructure initialization:** Creates DB clients, gRPC/HTTP servers, MQ clients — each with logging, metrics, and health check integration.
- **Health checks:** Registers DB and MQ with `HealthcheckRegistry`.
- **Lifecycle management:** Implements `app.Runner`/`app.Closer`; runners registered in dependency order, closers in reverse.
- **Remote config hot-reload:** `ReceiveConfig` upgrades infrastructure clients (DB, MQ, gRPC) with new parameters at runtime. Critical failures call `boot.Fatal()`.

#### `locator.go`
- **Dependency injection:** `Locator` holds minimal infra deps (DB, logger) and provides `Handlers(conf)` to construct the full application layer graph.
- **Layered wiring (dependency order):**
  1. **Repository:** Creates repos, injecting DB interface.
  2. **Service:** Creates services, injecting repo interfaces (local interface types).
  3. **Controller:** Creates controllers, injecting service interfaces (local interface types).
- **Transport handlers:** Wraps controllers with middleware/logging into gRPC mux, HTTP router, and RabbitMQ consumers (Ack/Retry/DLQ).

---

### `conf` Package

**Location:** `conf/`

**Responsibility:** Configuration structures with local and remote mechanisms.

- **Remote config struct (`conf.Remote`):** Fetched from `isp-config-service`. Contains DB params, MQ settings, log level, consumer config and so on. Supports JSON schema validation.
- **Local config files:** `config.yml`, `config_dev.yml` — static YAML with config service address, gRPC bind addresses, module name, log rotation, metrics autodiscovery. `config_dev.yml` used when `APP_MODE=dev`.
- **Remote config template:** `default_remote_config.json` — default template sent to config service on first connection. Uses placeholders (e.g., `{{ msp_pgsql_address }}`).
- **Validation:** Tests validate default remote config against struct schema.

Remote config supports **hot-reload** via `ReceiveConfig` in assembly, enabling runtime changes without restart. Follows **12-factor app** principles — no secrets hardcoded.

---

### `domain` Package

**Location:** `domain/`

**Responsibility:** API contract layer — request/response DTOs and error codes between `controller` and `service`.

- **DTOs:** Structs for data transfer between layers. Distinct from `entity` (persistence models).
- **Error codes:** Numeric constants (e.g., `ErrCodeObjectNotNotFound = 800`) returned to clients.
- **Validation tags:** `validate:"required"`, `validate:"required,max=32"` — enforced at transport layer before reaching service.

Contains only data structures and error codes — no business logic, interfaces, or infrastructure.

---

### `entity` Package

**Location:** `entity/`

**Responsibility:** Domain entities — core data model as stored in the database.

- **Entity structs:** Core domain data structures (persistence models).
- **Custom types:** Implement `sql.Scanner`/`driver.Valuer` for DB serialization (e.g., PostgreSQL JSON columns).
- **Sentinel errors:** Domain-specific error values (e.g., `ErrObjectNotNotFound`, `ErrMessageNotFound`) checked via `errors.Is()` across layers.

Entities are independent of storage technology or transport format. The `entity` layer sits at the **center of the onion** — both `repository` and `service` depend on it, but it depends on neither.

---

### `repository` Package

**Location:** `repository/`

**Responsibility:** Data access layer. Implements interfaces defined by `service`. Encapsulates all external system interactions (databases, caches, message brokers, HTTP APIs, file storage, Kafka, Redis).

- **Repository structs:** Wrap external system clients (DB connections, HTTP clients, cache, producers). Methods accept `context.Context` for cancellation/tracing.
- **Error translation:** Translates low-level errors (e.g., `sql.ErrNoRows`, HTTP 404) into domain sentinel errors (e.g., `entity.ErrObjectNotNotFound`). Other errors wrapped with `errors.WithMessage`.
- **Observability:** All operations annotated with operation labels (e.g., `sql_metrics.OperationLabelToContext`).

Repositories depend only on `entity` — never on `service` or `controller`. Different implementations can coexist (SQL repo, HTTP gateway, Redis cache), each wrapping its respective client but conforming to the same interface pattern.

---

### `service` Package

**Location:** `service/`

**Responsibility:** Business logic layer. Orchestrates repository calls, manages transactions, transforms between `entity` and `domain` types.

- **Service structs:** Hold dependencies as interface types defined locally (dependency inversion). Can depend on **other services** through locally-defined interfaces (service composition). The locator wires these inter-service dependencies.
- **Business logic methods:**
  - Call repository methods for data I/O.
  - Call other service interfaces for cross-domain logic.
  - Transform between `entity` and `domain` types.
  - Wrap errors with `errors.WithMessage`/`errors.WithMessagef`.
  - Check sentinel errors via `errors.Is()` for conditional logic (e.g., "if not found, insert; if newer, update").

The service layer is the **use case layer**. It imports only `entity` and `domain` — never `repository` or `controller`. Errors are wrapped with context but not translated to protocol-specific types; the error chain is preserved via `errors.Is()`.

---

### `controller` Package

**Location:** `controller/`

**Responsibility:** Transport adapter layer. Handles HTTP, gRPC, and message queue requests; performs validation and error translation; delegates to `service`.

- **Controller structs:** Hold a service interface (defined locally) as dependency.
- **Request handling:**
  - **HTTP/gRPC:** Accept request DTOs (from `domain`), delegate to service, return response DTOs. Validation tags enforced by transport layer before handler is called.
  - **Message handlers:** Accept delivery objects, unmarshal payload into entity structs, delegate to service, return Ack/Retry/DLQ.
- **Error translation:**
  - **gRPC/HTTP:** Maps sentinel errors to structured business errors with numeric codes (`apierrors.NewBusinessError`). Other errors → `apierrors.NewInternalServiceError`.
  - **Message queues:** Determines disposition — Ack (success), Retry (transient), MoveToDlq (permanent).
- **API documentation:** Swagger annotations on handler methods.

Controllers are **thin adapters** (anti-corruption layer) — no business logic, only protocol-specific handling. Error translation happens exclusively here, keeping the service layer transport-agnostic.

---

### `routes` Package

**Location:** `routes/`

**Responsibility:** Routing configuration for gRPC and HTTP. Maps endpoint paths to controller methods.

- **Controllers registry:** Struct aggregating all controller instances. New controllers added via struct fields.
- **Endpoint descriptors:** Functions returning `cluster.EndpointDescriptor` objects (path, HTTP method, auth, handler). Used by framework for service discovery.
- **Handler builders:** Construct transport-specific handlers:
  - **gRPC:** Builds mux by iterating endpoint descriptors, registering handlers with middleware.
  - **HTTP:** Builds router mapping URL paths to controller methods with middleware.
- **Middleware wrappers:** Uses `endpoint.DefaultWrapper`, `httpEndpoint.DefaultWrapper` for uniform logging/metrics/tracing.

Called from `main.go` during bootstrap (`routes.EndpointDescriptors()` provides cluster framework with service metadata).

---

## Error Handling Flow

Layered error handling where errors are progressively translated from infrastructure errors to protocol-specific responses.

1. **Repository:** Wraps errors with `errors.WithMessage`. Translates infra errors (e.g., `sql.ErrNoRows`) to domain sentinel errors (e.g., `entity.ErrObjectNotNotFound`). Returns sentinels checkable via `errors.Is()`.
2. **Service:** Wraps repository errors with context. Checks `errors.Is()` for conditional logic. Does NOT translate to protocol types — propagates with context, preserving error chain.
3. **Controller:** Final error translation by transport:
   - **gRPC/HTTP:** Sentinel errors → business errors with numeric codes; others → internal service errors.
   - **Message queues:** Deserialization errors → DLQ; service errors → Retry; success → Ack.

```
Infrastructure (sql.ErrNoRows, connection errors)
  → Repository: wrap + translate to sentinel errors
    → Service: wrap + errors.Is() for conditional logic
      → Controller: translate to protocol response
          gRPC/HTTP: Business error (code) / Internal error
          RMQ:       DLQ / Retry / Ack
```

---

## Dependency Flow

Strict layered architecture with unidirectional dependencies:

```
main
  └── assembly (composition root, infrastructure lifecycle)
       ├── conf (configuration structures)
       ├── locator (dependency injection / object graph)
       │    ├── repository → entity
       │    ├── service → entity, domain
       │    ├── controller → domain, entity
       │    ├── routes → controller
       │    └── transaction → repository, service
       └── conf
```

**Rules:**
- `main` → `assembly`, `conf`, `routes`
- `assembly` → `conf`, `repository`, `service`, `controller`, `routes`, `transaction`, `isp-kit`
- `repository` → `entity`, `isp-kit/db`
- `service` → `entity`, `domain`
- `controller` → `domain`, `entity`
- `routes` → `controller`
- `transaction` → `repository`, `service`

No circular dependencies. Dependencies always point inward (toward domain core). Inner layers define interfaces; outer layers provide implementations (**Clean Architecture** / **Onion Architecture**).

---