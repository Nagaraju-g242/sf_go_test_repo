# API type organization

Each Snowflake resource has its own type file:

- `role_types.go`
- `user_types.go`
- `database_types.go`
- `schema_types.go`
- `warehouse_types.go`
- `grant_types.go`
- `common_types.go` contains shared status.
- `groupversion.go` contains API group/version and scheme registration.

This keeps the API package easier to maintain while preserving the same Kubernetes
scheme registration and DeepCopy behavior.
