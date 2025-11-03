# Authorization

Thunder provides flexible authorization capabilities to control access to resources in your applications. Whether you're building fine-grained access control for web applications, mobile apps, or service-to-service integrations, Thunder offers role-based access control (RBAC) with multiple integration approaches to meet your needs.

## 🔒 Authorization Approaches

Thunder supports two main authorization approaches:

### 1. Flow-Based Authorization

Authorization integrated directly into authentication flows, where permissions are evaluated and included in authentication assertions.

[Learn more about Flow-Based Authorization →](./flow-based-authorization.md)

### 2. OAuth-Based Authorization with Custom Scopes

Standards-compliant OAuth 2.0 authorization with custom permission scopes that extend beyond OpenID Connect scopes.

[Learn more about OAuth Scopes and Permissions →](./oauth-scopes-permissions.md)

## 🎯 Role-Based Access Control (RBAC)

Both authorization approaches rely on Thunder's Role-Based Access Control system:

- **Roles** - Define sets of permissions (e.g., "DocumentEditor", "Admin")
- **Permissions** - Granular access rights (e.g., "read:documents", "write:documents")
- **Assignments** - Associate roles with users or groups
- **Organization Units** - Scope roles to specific organizational contexts

[Learn more about Role-Based Access Control →](./role-based-access-control.md)
