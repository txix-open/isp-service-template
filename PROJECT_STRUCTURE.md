# Project Structure Documentation

## Overview

`isp-service-template` is a Go microservice template built on the `txix-open/isp-kit` framework. It implements **Clean Architecture** principles, providing a standardized structure where each layer has a strictly defined responsibility. The template accelerates microservice development by handling infrastructure concerns (logging, configuration, tracing, database connections, message queues), allowing developers to focus on business logic.

The module name is `isp-service-template` and it depends on `github.com/txix-open/isp-kit v1.72.0`, which provides the bootstrap, cluster communication, database access (`dbrx`, `dbx`), gRPC/HTTP servers, RabbitMQ integration (`grmqx`), observability, and testing utilities.

---

## Package Responsibilities

### Root Module (`main.go`)

**Location:** `main.go`

**Responsibility:** The entry point of the application. It bootstraps the service using `isp-kit/bootstrap`, wires up the assembly layer, registers lifecycle runners and closers, and starts the application server.

**Key workflow:**
1. Creates a `bootstrap.Bootstrap` instance via `bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors(), cluster.GrpcTransport)`. This initializes configuration loading, logging, Sentry integration, health checks, and cluster communication.
2. Instantiates the `assembly.Assembly` by calling `assembly.New(boot)`, which sets up database clients, gRPC/HTTP servers, MDM client, and RabbitMQ client.
3. Registers assembly runners (gRPC server listen/serve, cluster client) and closers (cluster, gRPC/HTTP shutdown, MQ close, DB close, MDM close) with the app.
4. Registers a shutdown handler via `shutdown.On()` that gracefully shuts down the app on signal.
5. Calls `app.Run()` to start the application.

The `version` variable (`"1.0.0"`) is passed to bootstrap for identification. Swagger annotations (`@title`, `@version`, `@description`, `@host`, `@BasePath`) are present for API documentation generation via `swag init`.

**Philosophy:** `main.go` is intentionally thin. It delegates all infrastructure and business-logic initialization to the `assembly` package, following the **dependency inversion principle** — the entry point depends on abstractions (interfaces), not concrete implementations.

---

### `assembly` Package

**Location:** `assembly/`

**Responsibility:** The assembly package is the composition root of the application. It is responsible for wiring together all infrastructure components and dependencies. It bridges the gap between the framework (`isp-kit/bootstrap`) and the application's internal layers (`repository`, `service`, `controller`).

**Philosophy:** The assembly package follows the **composition root pattern**. It contains no business logic — its sole purpose is to construct and configure objects, then inject them into the appropriate layers. This keeps infrastructure concerns isolated from domain logic and makes the application testable by allowing easy substitution of real dependencies with test doubles.

#### `assembly.go`

**File responsibility:** Initializes and configures all infrastructure clients and servers, manages the application lifecycle, and handles remote configuration updates.

**Key components:**
- `Assembly` struct holds references to: `bootstrap.Bootstrap`, `dbrx.Client` (database), `grpc.Server`, `http.Server`, MDM client, logger, and `grmqx.Client` (RabbitMQ).
- `New(boot)` — Constructor that:
  - Creates the database client (`dbrx.New`) with a migration runner.
  - Creates the MDM client (`client.Default()`).
  - Creates the RabbitMQ client (`grmqx.New`) wrapped with Sentry error logging.
  - Registers health checks for `db` and `mq`.
  - Initializes gRPC server (`grpc.DefaultServer()`), HTTP server (`http.NewServer(logger)`), and returns the `Assembly`.
- `ReceiveConfig(ctx, remoteConfig)` — Hot-reloads remote configuration:
  - Upgrades the remote config using `rc.Upgrade[conf.Remote]`, which merges new config with the existing one using a TTL-based diff mechanism.
  - Updates the log level.
  - Upgrades the database client with new connection parameters.
  - Creates a new `Locator` and generates handlers from the updated config.
  - Upgrades the gRPC server with the new handler mux.
  - Upgrades the RabbitMQ client with new consumer configuration.
  - Upgrades the HTTP server with the new handler.
  - On any critical error during config upgrade, calls `boot.Fatal()` to terminate the application.
- `Runners()` — Returns lifecycle runners:
  - gRPC server's `ListenAndServe` on the bootstrap binding address.
  - Cluster client's `Run` (for service discovery and remote config subscription).
  - HTTP server runner is commented out (disabled by default).
- `Closers()` — Returns lifecycle closers in reverse order: cluster client, gRPC server shutdown, HTTP server shutdown, RabbitMQ close, database close, MDM client close.

**Philosophy:** The `Assembly` implements the `app.Runner` and `app.Closer` interfaces from `isp-kit/app`, providing a clean lifecycle management abstraction. The hot-reload mechanism (`ReceiveConfig`) allows the service to pick up configuration changes from `isp-config-service` without restarting, making it suitable for dynamic cloud environments.

#### `locator.go`

**File responsibility:** Implements dependency injection between the application layers — building the object graph from repositories through services to controllers, and wiring them into transport handlers (gRPC, HTTP, RabbitMQ).

**Key components:**
- `DB` interface — combines `db.DB` and `db.Transactional` interfaces, defining what the locator needs from the database layer.
- `Locator` struct — holds `db` and `logger`, the minimal dependencies needed to construct the entire application layer graph.
- `LocatorConfig` struct — the output of the locator: `HttpHandler` (HTTP router), `GrpcHandler` (gRPC mux), and `RmqHandler` (RabbitMQ consumer).
- `NewLocator(db, logger)` — Creates a `Locator` with the given database and logger.
- `Handlers(conf)` — The core method that builds the complete dependency graph:
  1. **Repository layer:** Creates `repository.NewObject(db)` — the object repository.
  2. **Service layer:** Creates `service.NewObject(objectRepo)` — the object service, which receives the repository interface.
  3. **Controller layer:** Creates `controller.NewObject(objectService)` — the object controller, which receives the service interface.
  4. **Routes assembly:** Packages controllers into `routes.Controllers{Object: objectController}`.
  5. **gRPC handler:** Creates an `endpoint.DefaultWrapper` with gRPC logging, then builds the gRPC mux via `routes.Handler(mapper, c)`.
  6. **HTTP handler:** Creates an `httpEndpoint.DefaultWrapper` with HTTP logging, then builds the HTTP router via `routes.HttpHandler(wrapper, c)`.
  7. **Transaction management:** Creates `transaction.NewManager(l.db)` — the transaction manager wrapping the database.
  8. **Message service:** Creates `service.NewMessage(logger, txManager)` — the message service, which receives the transaction manager (implementing `MessageTransactionRunner`).
  9. **Message controller:** Creates `controller.NewMessage(msgService)`.
  10. **RabbitMQ handler:** Wraps the message controller with `grmqx.NewResultHandler`, then creates the consumer via `conf.Consumer.Config.DefaultConsumer(handler, grmqx.ConsumerLog(...))`.
  11. Returns `LocatorConfig` with all three transport handlers.

**Philosophy:** The locator implements the **dependency injection pattern** at the application level. Each layer depends only on interfaces defined within its own package or the layer below it (e.g., `service` defines `Repo` interface, `controller` defines `ObjectService` interface). This ensures loose coupling — the controller doesn't know about the repository implementation, and the service doesn't know about the repository's concrete type. The locator is the only place that knows about concrete implementations, making it easy to swap dependencies for testing.

---

### `conf` Package

**Location:** `conf/`

**Responsibility:** Defines the application's configuration structures and provides both local and remote configuration mechanisms.

**Key files:**
- `remote.go` — Defines the `Remote` struct (remote configuration from `isp-config-service`) with `Database`, `Consumer`, and `LogLevel` fields. The `Consumer` struct wraps `grmqx.Connection` and `grmqx.Consumer`. An `init()` function registers a custom JSON schema generator for the `LogLevel` field, producing an enum of `debug`, `info`, `warn`, `error`, `fatal`.
- `remote_test.go` — Validates the `default_remote_config.json` against the `conf.Remote` struct using `rct.Test` (isp-kit's remote config testing utility).
- `default_remote_config.json` — Default remote configuration template sent to `isp-config-service`. Contains database connection parameters (templated with `{{ msp_pgsql_* }}` variables), log level (`debug`), and RabbitMQ consumer configuration (queue name, DLQ enabled, prefetch count, concurrency, retry policy with infinite retries).
- `config.yml` / `config_dev.yml` — Local configuration files. Define `configServiceAddress` (isp-config-service endpoint), `grpcOuterAddress`/`grpcInnerAddress` (gRPC bind/publish addresses), `moduleName`, `remoteConfigReceiverTimeout`, `logfile` (rotation settings), and `metricsAutodiscovery`.

**Philosophy:** Configuration is split into local (static, file-based) and remote (dynamic, fetched from `isp-config-service`). The remote config supports hot-reload via the `ReceiveConfig` mechanism, allowing runtime changes to database connections, log levels, and consumer settings without restarting the service.

---

### `domain` Package

**Location:** `domain/`

**Responsibility:** Defines request/response structures and error codes that flow between the `controller` and `service` layers. This is the API contract layer.

**Key files:**
- `object.go` — Defines `ErrCodeObjectNotFound` (error code `800`) and the `Object` struct with a `Name` field (with validation tags: `required`, `max=32`).
- `request.go` — Defines the `ByIdRequest` struct with an `Id` field (with validation tag: `required`).

**Philosophy:** The `domain` package contains only data structures and error codes — no business logic, no interfaces, no infrastructure concerns. It serves as the shared vocabulary between the transport layer (controller) and the business logic layer (service). Validation tags on struct fields enable automatic request validation at the controller/transport level.

---

### `entity` Package

**Location:** `entity/`

**Responsibility:** Defines domain entities — the core data model of the application. These structures represent data as stored in the database or passed between layers.

**Key files:**
- `object.go` — Defines the `Object` entity with `Id` (string) and `Name` (string) fields. This is the persistence model, distinct from the `domain.Object` which is the API-facing model.
- `message.go` — Defines the `Message` entity (`Id`, `Version`, `Data`) and `MessageData` struct (`Text` field). `MessageData` implements `sql.Scanner` and `driver.Valuer` interfaces for PostgreSQL JSON column serialization/deserialization using `isp-kit/json`.
- `errors.go` — Defines sentinel errors: `ErrObjectNotFound` and `ErrMessageNotFound`, used by the repository layer to signal "not found" conditions that the service/controller layers can check with `errors.Is()`.

**Philosophy:** Entities are the purest representation of the domain data. They are independent of any specific storage technology or transport format. The `entity` layer sits at the center of the onion — both `repository` and `service` depend on it, but it depends on neither. Errors defined here are sentinel errors that allow precise error checking across layers using `errors.Is()`.

---

### `repository` Package

**Location:** `repository/`

**Responsibility:** The data access layer. Implements interfaces defined by the `service` layer and encapsulates all interactions with external data sources (PostgreSQL database, advisory locks).

**Key files:**
- `object.go` — `Object` repository struct wrapping `db.DB`. Implements:
  - `All(ctx)` — SELECT all objects ordered by ID, returns `[]entity.Object`.
  - `Get(ctx, id)` — SELECT a single object by ID using Squirrel query builder. Returns `entity.ErrObjectNotFound` when `sql.ErrNoRows` is encountered.
- `message.go` — `Message` repository struct wrapping `db.DB`. Implements:
  - `Insert(ctx, msg)` — INSERT a new message record.
  - `GetLastVersion(ctx, id)` — SELECT the version of a message by ID. Returns `entity.ErrMessageNotFound` when not found.
  - `UpdateById(ctx, msg)` — UPDATE message by ID using Squirrel.
- `locker.go` — `Locker` struct wrapping `db.DB`. Implements PostgreSQL advisory locking:
  - `Lock(ctx, key)` — Acquires an exclusive transaction-level advisory lock using `pg_advisory_xact_lock`. Key is hashed with FNV-1a and prefixed with `"isp-service-template"`.
  - `TryLock(ctx, key)` — Attempts to acquire the lock non-blocking using `pg_try_advisory_xact_lock`.

**Philosophy:** Repositories implement the **repository pattern**. They depend only on the `db.DB` interface from `isp-kit/db` and the `entity` package — never on `service` or `controller`. All SQL operations are wrapped with `sql_metrics.OperationLabelToContext` for observability. Errors are wrapped with context using `errors.WithMessage`. The repository returns domain-specific sentinel errors (`entity.ErrObjectNotFound`, `entity.ErrMessageNotFound`) rather than raw database errors, allowing higher layers to handle them semantically.

---

### `service` Package

**Location:** `service/`

**Responsibility:** The business logic layer. Contains the core application logic, orchestrates repository calls, manages transactions, and transforms between `entity` (persistence model) and `domain` (API model) types.

**Key files:**
- `object.go` — `Object` service struct holding a `Repo` interface (defined locally: `All`, `Get`). Implements:
  - `All(ctx)` — Calls `repo.All()`, maps `[]entity.Object` to `[]domain.Object` (stripping the `Id` field, keeping only `Name`).
  - `Get(ctx, id)` — Calls `repo.Get()`, maps `entity.Object` to `domain.Object`. Errors are wrapped with `errors.WithMessagef`.
- `message.go` — `Message` service struct holding a `log.Logger` and a `MessageTransactionRunner` interface. Implements:
  - `Handle(ctx, msg)` — Wraps the entire message handling in a transaction via `txRunner.MessageTransaction()`. Inside the transaction:
    - Acquires a lock on the message ID.
    - Checks if a message with this ID exists (`GetLastVersion`).
    - If not found (`entity.ErrMessageNotFound`), inserts a new message.
    - If the incoming message's version is higher than the stored version, updates the record.
    - If the incoming message's version is lower or equal, skips the update (idempotency).
  - `handle(ctx, msg, tx)` — The inner function that performs the actual logic within the transaction, receiving a `MessageTransaction` interface (combining `Locker` and `Message` repository methods).

**Philosophy:** The service layer defines its own interfaces for dependencies (`Repo` in `object.go`, `MessageTransaction`/`MessageTransactionRunner` in `message.go`), following the **dependency inversion principle**. It never imports `repository` or `controller` — only `entity` and `domain`. The `Message` service demonstrates **transactional consistency**: all operations (lock, read, write) happen within a single database transaction, ensuring atomicity. The version-based update logic provides **idempotency** — duplicate or out-of-order messages are handled gracefully.

---

### `controller` Package

**Location:** `controller/`

**Responsibility:** The transport adapter layer. Handles incoming requests from HTTP, gRPC, and RabbitMQ, performs initial validation and error translation, and delegates to the `service` layer. Controllers are the boundary between external protocols and internal business logic.

**Key files:**
- `object.go` — `Object` controller struct holding an `ObjectService` interface (`All`, `Get`). Implements:
  - `All(ctx)` — Delegates directly to `service.All()`, returning `[]domain.Object`.
  - `GetById(ctx, req)` — Delegates to `service.Get()`, with **error translation**:
    - If the error is `entity.ErrObjectNotFound`, translates it to a gRPC business error with code `domain.ErrCodeObjectNotFound` (800) and a descriptive message.
    - For any other error, wraps it as a gRPC internal service error via `apierrors.NewInternalServiceError`.
  - Swagger annotations document the API endpoints (tags, summaries, request/response schemas, error codes).
- `message.go` — `Message` controller struct holding a `MessageService` interface (`Handle`). Implements:
  - `Handle(ctx, delivery)` — RabbitMQ message handler:
    - Unmarshals the delivery body into `entity.Message`. On failure, returns `handler.MoveToDlq` (dead-letter queue).
    - Calls `service.Handle()`. On failure, returns `handler.Retry` (message will be retried per the retry policy). On success, returns `handler.Ack()`.

**Philosophy:** Controllers are **thin adapters**. They contain no business logic — only protocol-specific handling (JSON unmarshaling, error translation, HTTP/gRPC/RMQ response formatting). The controller defines interfaces for the services it depends on (`ObjectService`, `MessageService`), ensuring the transport layer is decoupled from the business logic implementation. Error handling at this layer translates internal domain errors into protocol-appropriate responses (gRPC `apierrors`, RMQ retry/DLQ decisions).

---

### `routes` Package

**Location:** `routes/`

**Responsibility:** Defines routing configuration for gRPC and HTTP transports, mapping endpoint paths to controller methods. Acts as the bridge between transport handlers and controllers.

**Key files:**
- `routes.go` — Defines:
  - `Controllers` struct — holds references to all controllers (currently `Object`).
  - `EndpointDescriptors()` — Returns a list of `cluster.EndpointDescriptor` for gRPC service discovery/registration. Each descriptor includes the path, auth requirements, HTTP method, and handler reference. Called from `main.go` during bootstrap.
  - `Handler(wrapper, c)` — Builds a gRPC mux (`grpc.Mux`) by iterating over endpoint descriptors and registering each handler with the provided wrapper.
  - `HttpHandler(wrapper, c)` — Builds an HTTP router (`router.Router`) with POST routes for `/object/all` and `/object/get_by_id`, each wrapped with the provided HTTP endpoint wrapper.
  - `endpointDescriptors(c)` — Internal function returning the concrete endpoint list with paths like `"isp-service-template/object/all"` and `"isp-service-template/object/get_by_id"`.

**Philosophy:** The routes package is the **transport configuration layer**. It knows about URL paths and HTTP methods but contains no business logic. The `Controllers` struct acts as a registry for all available controllers, making it easy to add new endpoints. The `EndpointDescriptors()` function (called from `main.go`) provides the cluster with service metadata for discovery. The wrapper pattern (`endpoint.Wrapper`, `httpEndpoint.Wrapper`) allows cross-cutting concerns (logging, metrics, tracing) to be applied uniformly.

---

### `transaction` Package

**Location:** `transaction/`

**Responsibility:** Provides transaction management utilities that wrap repository operations within database transactions.

**Key files:**
- `manager.go` — Defines:
  - `Manager` struct — holds a `db.Transactional` interface.
  - `NewManager(db)` — Constructor.
  - `messageTx` struct — embeds `repository.Locker` and `repository.Message`, combining them into a single `service.MessageTransaction` interface implementation.
  - `MessageTransaction(ctx, fn)` — Executes the provided function within a database transaction (`db.RunInTransaction`). Inside the transaction, creates a `Locker` and `Message` repository from the transaction handle (`*db.Tx`), wraps them in a `messageTx` struct, and passes it to the callback function.

**Philosophy:** The transaction manager implements the **transaction script pattern**. It provides a clean API for wrapping multi-step database operations in a single transaction, ensuring atomicity. By combining multiple repositories (`Locker` + `Message`) into a single `messageTx` struct, it allows the service layer to perform coordinated operations (lock + read + write) within one transaction. The service layer interacts with this through the `MessageTransactionRunner` interface, keeping it decoupled from the transaction implementation.

---

## Error Handling Flow

The project implements a layered error handling strategy where errors are progressively translated from low-level infrastructure errors to high-level protocol-specific responses:

### Layer-by-Layer Error Flow

1. **Repository layer** (`repository/`):
   - Wraps all errors with context using `errors.WithMessage(err, "operation description")`.
   - Translates `sql.ErrNoRows` into domain-specific sentinel errors (`entity.ErrObjectNotFound`, `entity.ErrMessageNotFound`).
   - Example: `repository/object.go:54` — `if errors.Is(err, sql.ErrNoRows) { return nil, entity.ErrObjectNotFound }`

2. **Service layer** (`service/`):
   - Wraps repository errors with additional context using `errors.WithMessage` or `errors.WithMessagef`.
   - Checks for sentinel errors using `errors.Is()` to implement conditional logic (e.g., `service/message.go:53` — `if errors.Is(err, entity.ErrMessageNotFound)`).
   - Does NOT translate errors to protocol-specific types — it propagates them up with context.

3. **Controller layer** (`controller/`):
   - Performs the final error translation based on the transport protocol.
   - **Object controller** (`controller/object.go`):
     - Checks for `entity.ErrObjectNotFound` and converts it to a gRPC business error: `apierrors.NewBusinessError(domain.ErrCodeObjectNotFound, "message", err)` — this produces a structured gRPC error with a business error code (800) that clients can programmatically handle.
     - All other errors are wrapped as internal service errors: `apierrors.NewInternalServiceError(err)`.
   - **Message controller** (`controller/message.go`):
     - JSON unmarshal errors result in `handler.MoveToDlq` (message goes to dead-letter queue).
     - Service errors result in `handler.Retry` (message will be retried per the configured retry policy).
     - Success results in `handler.Ack()`.

### Error Flow Summary

```
Database (sql.ErrNoRows)
  → Repository wraps with context + translates to entity.ErrObjectNotFound
    → Service wraps with context, checks errors.Is for conditional logic
      → Controller translates to protocol-specific response:
          gRPC: apierrors.NewBusinessError (code 800) or apierrors.NewInternalServiceError
          RMQ:  handler.MoveToDlq / handler.Retry / handler.Ack
```

**Key design principles:**
- Errors are **wrapped with context** at each layer, preserving the original cause for debugging via `errors.Is()` and `errors.As()`.
- **Sentinel errors** (`entity.ErrObjectNotFound`, `entity.ErrMessageNotFound`) are defined at the `entity` layer, allowing any layer to check for them without importing infrastructure packages.
- **Protocol-specific translation** happens only at the controller layer, keeping the service layer transport-agnostic.
- The gRPC wrapper (`endpoint.DefaultWrapper` with `grpclog.Log`) and HTTP wrapper (`httpEndpoint.DefaultWrapper` with `httplog.Log`) from `isp-kit` handle logging and serialization of these errors into the appropriate wire format.

---

## Request/Response Flow

### HTTP/gRPC Request Flow (Object operations)

```
HTTP/gRPC request
  → routes (path → controller method)
    → controller (validation, error translation, delegates to service)
      → service (business logic, calls repository, transforms entity ↔ domain)
        → repository (SQL queries, returns entity + sentinel errors)
          → database
```

### RabbitMQ Message Flow (Message processing)

```
RabbitMQ message
  → controller.Handle (JSON unmarshal, delegates to service)
    → service.Handle (transaction wrapper, calls transaction manager)
      → transaction.Manager.MessageTransaction (opens DB transaction)
        → service.handle (lock → check version → insert/update)
          → repository (Locker + Message operations within transaction)
            → database
```

---

## Dependency Graph

```
main
  └── assembly
       ├── conf (configuration structures)
       ├── repository (data access)
       │    └── entity (data models, sentinel errors)
       ├── service (business logic)
       │    ├── entity (data models)
       │    └── domain (request/response models)
       ├── controller (transport adapters)
       │    ├── domain (request/response models)
       │    └── entity (sentinel errors)
       ├── routes (routing configuration)
       │    └── controller (controller interfaces)
       └── transaction (transaction management)
            ├── repository (Locker, Message)
            └── service (MessageTransaction interface)
```

**Key dependency rules:**
- `main` depends on `assembly`, `conf`, `routes`.
- `assembly` depends on `conf`, `repository`, `service`, `controller`, `routes`, `transaction`.
- `repository` depends only on `entity` and `isp-kit/db`.
- `service` depends on `entity` and `domain`.
- `controller` depends on `domain` and `entity`.
- `routes` depends on `controller`.
- `transaction` depends on `repository` and `service`.
- No circular dependencies exist between application packages.