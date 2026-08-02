# Project Structure Documentation

## Overview

This is a Go microservice template built on the `txix-open/isp-kit` framework. It implements **Clean Architecture** principles, providing a standardized structure where each layer has a strictly defined responsibility. The template accelerates microservice development by handling infrastructure concerns (logging, configuration, tracing, database connections, message queues), allowing developers to focus on business logic.

The framework (`isp-kit`) provides bootstrap, cluster communication, database access (`dbrx`, `dbx`), gRPC/HTTP servers, RabbitMQ integration (`grmqx`), observability, and testing utilities.

---

## Package Responsibilities

### Root Module (`main.go`)

**Location:** `main.go`

**Responsibility:** The entry point of the application. It bootstraps the service using `isp-kit/bootstrap`, wires up the assembly layer, registers lifecycle runners and closers, and starts the application server.

**Key workflow:**
1. Creates a `bootstrap.Bootstrap` instance via `bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors(), cluster.GrpcTransport)`. This initializes configuration loading, logging, Sentry integration, health checks, and cluster communication.
2. Instantiates the `assembly.Assembly` by calling `assembly.New(boot)`.
3. Registers assembly runners (gRPC server listen/serve, cluster client) and closers (cluster, gRPC/HTTP shutdown, MQ close, DB close) with the app.
4. Registers a shutdown handler via `shutdown.On()` that gracefully shuts down the app on signal.
5. Calls `app.Run()` to start the application.

**Philosophy:** `main.go` is intentionally thin. It delegates all infrastructure and business-logic initialization to the `assembly` package. The bootstrap framework handles configuration loading, logging setup, and cluster membership, while the assembly handles dependency wiring.

---

### `assembly` Package

**Location:** `assembly/`

**Responsibility:** The composition root of the application. It is responsible for wiring together all infrastructure components and dependencies. It bridges the gap between the framework (`isp-kit/bootstrap`) and the application's internal layers (`repository`, `service`, `controller`).

**Philosophy:** The assembly package follows the **composition root pattern**. It contains no business logic — its sole purpose is to construct and configure objects, then inject them into the appropriate layers. This keeps infrastructure concerns isolated from domain logic and makes the application testable by allowing easy substitution of real dependencies with test doubles.

#### `assembly.go`

**File responsibility:** Infrastructure initialization and lifecycle management. Initializes and configures all infrastructure clients and servers, manages the application lifecycle, and handles remote configuration updates.

**Key responsibilities:**
- **Infrastructure client initialization:** Creates database clients, gRPC/HTTP servers and message queue clients and so on. Each client is configured with appropriate logging, metrics, and health check integration.
- **Health check registration:** Registers infrastructure components (database, message queue) with the `HealthcheckRegistry` for liveness/readiness probes.
- **Lifecycle management:** Implements `app.Runner` and `app.Closer` interfaces to manage startup and graceful shutdown of all infrastructure components. Runners are registered in dependency order; closers are registered in reverse order.
- **Remote configuration hot-reload:** The `ReceiveConfig` method handles dynamic configuration updates pushed from a remote config service. It upgrades infrastructure clients (database, message queue, gRPC server) with new parameters without restarting the application. This enables runtime reconfiguration of connection strings, log levels, and consumer settings.
- **Error handling:** Critical failures during initialization or config upgrade call `boot.Fatal()` to terminate the application immediately.

**Philosophy:** The `Assembly` struct encapsulates all infrastructure state. By implementing `app.Runner`/`app.Closer`, it integrates cleanly with the framework's lifecycle management. The hot-reload mechanism demonstrates the **configuration-as-a-service** pattern, where configuration is dynamically pushed from a central config service rather than being statically loaded at startup.

#### `locator.go`

**File responsibility:** Dependency injection and object graph construction. Implements dependency injection between the application layers — building the object graph from repositories through services to controllers, and wiring them into transport handlers (gRPC, HTTP, RabbitMQ).

**Key responsibilities:**
- **Dependency injection container:** The `Locator` struct holds minimal infrastructure dependencies (database, logger) and provides a `Handlers(conf)` method that constructs the entire application layer graph.
- **Layered dependency wiring:** Constructs objects in dependency order:
  1. **Repository layer:** Creates repository instances, injecting the database interface.
  2. **Service layer:** Creates service instances, injecting repository interfaces (defined as local interface types).
  3. **Controller layer:** Creates controller instances, injecting service interfaces (defined as local interface types).
- **Transport handler assembly:** Wraps controllers with middleware/logging and registers them with transport-specific handlers:
  - gRPC: Creates an endpoint wrapper with gRPC logging, builds a gRPC mux mapping paths to controller methods.
  - HTTP: Creates an HTTP endpoint wrapper with HTTP logging, builds a router mapping URL paths to controller methods.
  - RabbitMQ: Wraps message controllers with result handlers (Ack/Retry/DLQ) and consumer logging.
- **Transaction management integration:** Creates transaction managers and injects them into services that require transactional consistency.

**Philosophy:** The locator implements the **dependency injection pattern** at the application level. Each layer defines its own interfaces for dependencies it needs (e.g., `service` defines `Repo` interface, `controller` defines `Service` interface). This ensures **loose coupling** — higher layers don't know about concrete implementations in lower layers. The locator is the only place that knows about concrete implementations, making it easy to swap dependencies for testing or to add new implementations.

The separation of `assembly.go` (infrastructure) and `locator.go` (application DI) reflects the **separation of concerns** principle: infrastructure lifecycle management is distinct from application object graph construction.

---

### `conf` Package

**Location:** `conf/`

**Responsibility:** Defines the application's configuration structures and provides both local and remote configuration mechanisms.

**Key components:**
- **Remote configuration struct:** Defines the structure for configuration fetched from a remote config service (`isp-config-service`). Contains database connection parameters, message broker settings, log level, and consumer configuration. Custom JSON schema generators can be registered for fields requiring special validation (e.g., enum constraints for log levels).
- **Local configuration files:** YAML files (`config.yml`, `config_dev.yml`) contain static local configuration such as the config service address, gRPC bind/publish addresses, module name, log file rotation settings, and metrics autodiscovery settings.
- **Remote config template:** A JSON file (`default_remote_config.json`) serves as the default configuration template sent to the config service on first connection. Connection parameters are templated using environment variable placeholders.
- **Configuration validation:** Tests validate that the default remote config matches the expected struct schema.

**Philosophy:** Configuration is split into local (static, file-based, environment-specific) and remote (dynamic, fetched from a config service). The remote config supports **hot-reload** via the `ReceiveConfig` mechanism in the assembly layer, allowing runtime changes to database connections, log levels, and consumer settings without restarting the service. This follows the **externalized configuration** pattern, where configuration is externalized from the application code and can be changed at runtime.

---

### `domain` Package

**Location:** `domain/`

**Responsibility:** Defines request/response structures and error codes that flow between the `controller` and `service` layers. This is the API contract layer.

**Key components:**
- **Request/Response DTOs:** Structs representing the data transferred between layers. These are distinct from `entity` structs, which represent persistence models.
- **Error codes:** Numeric constants representing business error codes (e.g., `ErrCodeObjectNotFound = 800`) that are returned to clients when specific business conditions occur.
- **Validation tags:** Struct fields include validation tags (e.g., `validate:"required"`, `validate:"required,max=32"`) that enable automatic request validation at the transport layer.

**Philosophy:** The `domain` package contains only data structures and error codes — no business logic, no interfaces, no infrastructure concerns. It serves as the **shared vocabulary** between the transport layer (controller) and the business logic layer (service). The separation between `domain` (API-facing) and `entity` (persistence-facing) models follows the **DTO pattern**, allowing each layer to have its own optimal representation of data. Validation tags at this level enable **fail-fast validation** before requests reach the service layer.

---

### `entity` Package

**Location:** `entity/`

**Responsibility:** Defines domain entities — the core data model of the application. These structures represent data as stored in the database or passed between layers.

**Key components:**
- **Entity structs:** Core domain data structures. These are the persistence models used by repositories.
- **Custom types with serialization:** Structs that implement `sql.Scanner` and `driver.Valuer` interfaces for database serialization/deserialization (e.g., JSON column types in PostgreSQL).
- **Sentinel errors:** Domain-specific error values (e.g., `ErrObjectNotFound`, `ErrMessageNotFound`) that represent well-known failure conditions. These are checked using `errors.Is()` across layers.

**Philosophy:** Entities are the purest representation of the domain data. They are independent of any specific storage technology or transport format. The `entity` layer sits at the **center of the onion** — both `repository` and `service` depend on it, but it depends on neither. This is the **domain model** in Clean Architecture terms.

Sentinel errors defined at this layer allow precise error checking across all layers using `errors.Is()`, without requiring higher layers to import infrastructure packages. This enables the **error wrapping pattern**: lower layers wrap errors with context while preserving the ability to check for specific sentinel errors at higher layers.

---

### `repository` Package

**Location:** `repository/`

**Responsibility:** The data access layer. Implements interfaces defined by the `service` layer and encapsulates all interactions with external systems — databases, caches, message brokers, external HTTP APIs, file storage, Kafka producers, Redis clients, or any other I/O boundary.

**Key components:**
- **Repository structs:** Each struct wraps an external system client (database connection, HTTP client, cache client, message producer, etc.) and provides methods for data operations (CRUD, queries, locks, publishes). Methods accept `context.Context` as the first parameter for cancellation and tracing.
- **Error translation:** Translates low-level external system errors (e.g., `sql.ErrNoRows`, HTTP 404, connection timeouts) into domain-specific sentinel errors (e.g., `entity.ErrObjectNotFound`). All other errors are wrapped with context using `errors.WithMessage`.
- **Observability integration:** All external system operations are annotated with operation labels (e.g., `sql_metrics.OperationLabelToContext`), enabling per-operation metrics and tracing.

**Philosophy:** Repositories implement the **repository pattern** (for data stores) and the **gateway pattern** (for external service clients) as described in Domain-Driven Design and Clean Architecture. They depend only on the `entity` package and the framework's client interfaces — never on `service` or `controller`. This ensures the data access layer is completely decoupled from business logic and can be swapped (e.g., PostgreSQL → MySQL, REST API → gRPC client) without affecting the service layer.

The repository returns domain-specific sentinel errors rather than raw infrastructure errors, allowing higher layers to handle errors semantically. All operations include observability instrumentation (metrics labels) for production monitoring. The use of query builders and typed clients prevents injection attacks and enables safe dynamic query/request construction.

Different repository implementations can coexist: a SQL repository for database access, an HTTP gateway for external API calls, a Redis repository for caching — each wrapping its respective client but all conforming to the same interface pattern defined by the service layer.

---

### `service` Package

**Location:** `service/`

**Responsibility:** The business logic layer. Contains the core application logic, orchestrates repository calls, manages transactions, and transforms between `entity` (persistence model) and `domain` (API model) types.

**Key components:**
- **Service structs:** Each struct holds dependencies (repositories, loggers, transaction managers, and **other services**) as interface types defined locally within the service file. This follows the **dependency inversion principle** — the service defines what it needs, not who provides it.
- **Cross-service dependencies:** Services can depend on other services through locally-defined interfaces. For example, a `Message` service might depend on an `Object` service interface to validate references or trigger side effects. This enables **service composition** — complex business operations can orchestrate multiple services while maintaining loose coupling. The locator (in the assembly layer) is responsible for wiring these inter-service dependencies.
- **Business logic methods:** Implement application use cases. These methods:
  - Call repository methods to read/write data.
  - Call methods on other service interfaces when cross-domain logic is needed.
  - Transform between `entity` and `domain` types (e.g., stripping internal fields, combining data from multiple sources).
  - Wrap errors with context using `errors.WithMessage` or `errors.WithMessagef`.
  - Check for sentinel errors using `errors.Is()` to implement conditional logic (e.g., "if not found, insert; if found and newer, update").
- **Transaction management:** Services that require multi-step database operations receive a transaction runner interface. They wrap their logic in a transaction callback, ensuring atomicity of related operations.

**Philosophy:** The service layer is the **use case layer** in Clean Architecture. It defines its own interfaces for dependencies, never importing `repository` or `controller` — only `entity` and `domain`. This ensures the business logic is completely independent of infrastructure and transport concerns.

Services can depend on **each other** through interfaces defined at the service level. When service A needs functionality from service B, it defines a local interface specifying the methods it requires from B. The locator then injects B's implementation into A. This creates a **service composition graph** where complex workflows are built by orchestrating simpler services, while each service remains independently testable with mock dependencies.

Error handling in the service layer follows the **error wrapping pattern**: errors from lower layers are wrapped with additional context but not translated to protocol-specific types. The service preserves the original error chain via `errors.Is()` compatibility, allowing higher layers to make translation decisions.

---

### `controller` Package

**Location:** `controller/`

**Responsibility:** The transport adapter layer. Handles incoming requests from HTTP, gRPC, and message queues, performs initial validation and error translation, and delegates to the `service` layer. Controllers are the boundary between external protocols and internal business logic.

**Key components:**
- **Controller structs:** Each struct holds a service interface (defined locally) as its dependency. The controller defines what service capabilities it needs, not which concrete service it uses.
- **Request handling methods:** Methods that correspond to API endpoints or message handlers:
  - **HTTP/gRPC handlers:** Accept request DTOs (from `domain`), delegate to service methods, and return response DTOs. Validation tags on request structs are enforced by the transport layer before the handler is called.
  - **Message handlers:** Accept message delivery objects, unmarshal the payload into entity structs, delegate to service methods, and return protocol-specific results (Ack/Retry/DLQ).
- **Error translation:** Converts internal errors into protocol-appropriate responses:
  - **gRPC/HTTP:** Translates sentinel errors (e.g., `entity.ErrObjectNotFound`) into structured business errors with numeric error codes (e.g., `domain.ErrCodeObjectNotfound`). Other errors become internal service errors.
  - **Message queues:** Determines message disposition — Ack (success), Retry (transient failure), or MoveToDlq (permanent failure like deserialization error).
- **API documentation:** Swagger annotations on handler methods define the API contract (tags, summaries, request/response schemas, error codes).

**Philosophy:** Controllers are **thin adapters** implementing the **anti-corruption layer pattern**. They contain no business logic — only protocol-specific handling (JSON unmarshaling, error translation, response formatting). By defining their own service interfaces, controllers are decoupled from service implementations.

Error translation happens exclusively at this layer, keeping the service layer transport-agnostic. The controller maps internal domain errors to protocol-appropriate responses: gRPC business errors with numeric codes for client-side handling, or message queue dispositions (Ack/Retry/DLQ) for asynchronous processing.

---

### `routes` Package

**Location:** `routes/`

**Responsibility:** Defines routing configuration for gRPC and HTTP transports, mapping endpoint paths to controller methods. Acts as the bridge between transport handlers and controllers.

**Key components:**
- **Controllers registry:** A struct that aggregates all controller instances, serving as a registry for available endpoints. New controllers are added by adding fields to this struct.
- **Endpoint descriptors:** Functions that return lists of `cluster.EndpointDescriptor` objects, each describing a endpoint (path, HTTP method, auth requirements, handler reference). These descriptors are used by the framework for service discovery and cluster communication.
- **Handler builders:** Functions that construct transport-specific handler objects:
  - **gRPC handler:** Builds a gRPC mux by iterating over endpoint descriptors and registering each handler with a middleware wrapper.
  - **HTTP handler:** Builds an HTTP router by mapping URL paths to controller methods, each wrapped with HTTP-specific middleware.
- **Middleware wrappers:** Uses framework-provided wrappers (`endpoint.DefaultWrapper`, `httpEndpoint.DefaultWrapper`) that apply cross-cutting concerns (logging, metrics, tracing) uniformly to all endpoints.

**Philosophy:** The routes package is the **transport configuration layer**. It knows about URL paths, HTTP methods, and gRPC service paths but contains no business logic. The `Controllers` struct acts as a **service registry**, making it easy to discover and add new endpoints.

The wrapper pattern allows cross-cutting concerns to be applied uniformly without polluting controller logic. The `EndpointDescriptors()` function (called from `main.go` during bootstrap) provides the cluster framework with service metadata for discovery and routing.

---

## Error Handling Flow

The project implements a layered error handling strategy where errors are progressively translated from low-level infrastructure errors to high-level protocol-specific responses. This follows the **error wrapping and translation pattern**.

### Layer-by-Layer Error Flow

1. **Repository layer:**
   - Wraps all errors with context using `errors.WithMessage(err, "operation description")`.
   - Translates infrastructure-specific errors (e.g., `sql.ErrNoRows`) into domain-specific **sentinel errors** (e.g., `entity.ErrObjectNotFound`, `entity.ErrMessageNotFound`).
   - Returns sentinel errors that higher layers can check with `errors.Is()`.

2. **Service layer:**
   - Wraps repository errors with additional context using `errors.WithMessage` or `errors.WithMessagef`.
   - Checks for sentinel errors using `errors.Is()` to implement conditional business logic (e.g., "if not found, insert new record; if version is newer, update").
   - Does NOT translate errors to protocol-specific types — it propagates them up with context, preserving the error chain.

3. **Controller layer:**
   - Performs the final **error translation** based on the transport protocol:
   - **gRPC/HTTP controllers:** Maps sentinel errors to structured business errors with numeric error codes (e.g., `apierrors.NewBusinessError(code, message, err)`). All other errors become internal service errors (`apierrors.NewInternalServiceError(err)`).
   - **Message queue controllers:** Determines message disposition:
     - Deserialization errors → Move to dead-letter queue (DLQ).
     - Service errors → Retry (per configured retry policy).
     - Success → Acknowledge (Ack).

### Error Flow Summary

```
Infrastructure (sql.ErrNoRows, connection errors)
  → Repository wraps with context + translates to sentinel errors
    → Service wraps with context, checks errors.Is for conditional logic
      → Controller translates to protocol-specific response:
          gRPC/HTTP: Business error (with code) or Internal error
          RMQ:       DLQ / Retry / Ack
```

### Key Design Principles

- **Error wrapping with context:** Errors are wrapped with descriptive context at each layer using `errors.WithMessage`, preserving the original cause for debugging via `errors.Is()` and `errors.As()`.
- **Sentinel errors at the entity layer:** Well-known error conditions are defined as sentinel errors in the `entity` package, allowing any layer to check for them without importing infrastructure packages.
- **Protocol-specific translation at the boundary:** Error-to-protocol translation happens only at the controller layer, keeping the service layer transport-agnostic and reusable across different transport mechanisms.
- **Framework handling of serialization:** The gRPC/HTTP wrappers (`endpoint.DefaultWrapper`, `httpEndpoint.DefaultWrapper`) from `isp-kit` handle logging and serialization of translated errors into the appropriate wire format.

---

## Dependency Flow

The project follows a strict layered architecture with unidirectional dependencies:

```
main
  └── assembly (composition root, infrastructure lifecycle)
       ├── conf (configuration structures)
       ├── locator (dependency injection / object graph)
       │    ├── repository → entity (data access, persistence models)
       │    ├── service → entity, domain (business logic)
       │    ├── controller → domain, entity (transport adapters)
       │    ├── routes → controller (routing configuration)
       │    └── transaction → repository, service (transaction management)
       └── conf (configuration)
```

**Dependency rules:**
- `main` depends on `assembly`, `conf`, `routes`.
- `assembly` depends on `conf`, `repository`, `service`, `controller`, `routes`, `transaction`, and `isp-kit` packages.
- `repository` depends only on `entity` and `isp-kit/db`.
- `service` depends on `entity` and `domain`.
- `controller` depends on `domain` and `entity`.
- `routes` depends on `controller`.
- `transaction` depends on `repository` and `service`.
- No circular dependencies exist between application packages.

The dependency direction is always inward (toward the domain core). Infrastructure and transport layers depend on the domain, but the domain never depends on them. This is the **Clean Architecture** / **Onion Architecture** principle: the inner layers define interfaces, and the outer layers provide implementations.

---

## Testing Strategy

The project includes integration tests in the `tests` package that exercise the full stack (database, transport, business logic) using test doubles provided by `isp-kit/test`:

- **HTTP tests:** Use `httpt.TestServer` to spin up a test HTTP server with the real handler chain, and `dbt.New` for a test database with migrations.
- **gRPC tests:** Use `grpct.TestServer` to spin up a test gRPC server with the real handler chain.
- **Message queue tests:** Use `grmqt.New` to spin up a test RabbitMQ instance and publish/consume messages.
- **Configuration tests:** Use `rct.Test` to validate remote configuration schemas.

Tests construct the full dependency graph using `assembly.NewLocator`, ensuring that the same wiring used in production is exercised in tests. This validates the integration between all layers while maintaining test simplicity.

---

## Configuration Strategy

The project uses a **dual configuration model**:

1. **Local configuration** (`config.yml`, `config_dev.yml`): Static YAML files loaded at startup. Contains environment-specific settings like config service address, gRPC ports, log file paths, and metrics settings. The `config_dev.yml` variant is used when `APP_MODE=dev`.

2. **Remote configuration** (`default_remote_config.json` + `conf.Remote` struct): Dynamic configuration fetched from `isp-config-service`. Contains database credentials, message broker settings, log level, and consumer configuration. Supports **hot-reload** — when the remote config service pushes updates, the `ReceiveConfig` method in the assembly layer upgrades all infrastructure clients (database, MQ, gRPC server) without restarting the application.

Connection parameters in the remote config template use environment variable placeholders (e.g., `{{ msp_pgsql_address }}`), which are resolved by the config service at deployment time. This enables **12-factor app** compliance — configuration is externalized and environment-specific, with no secrets hardcoded in the application.