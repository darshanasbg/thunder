# Role-Based Access Control (RBAC)

Thunder's Role-Based Access Control (RBAC) system provides flexible permission management through roles. Roles are collections of permissions that can be assigned to users or groups, enabling fine-grained access control across your applications.

## 🎯 Prerequisites

Before setting up RBAC, ensure you have:

1. **Organization Unit** - A container for your users and roles
2. **Users or Groups** - Entities to assign roles to
3. **Permission Model** - Defined permissions for your application resources

## 📝 Creating Roles

Roles define collections of permissions that can be assigned to users or groups.

### Step 1: Create an Organization Unit

If you don't already have an organization unit, create one:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/organization-units \
-d '{
    "name": "Engineering",
    "description": "Engineering department",
    "handle": "engineering"
}'
```

Note the `id` from the response - you'll need this for creating roles and users.

### Step 2: Create a Role with Permissions

Create a role and assign permissions to it:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles \
-d '{
    "name": "DocumentEditor",
    "description": "Can read and write documents",
    "ouId": "<organization-unit-id>",
    "permissions": [
        "read:documents",
        "write:documents"
    ]
}'
```

**Response:**

```json
{
    "id": "role-123",
    "name": "DocumentEditor",
    "description": "Can read and write documents",
    "ouId": "ou-456",
    "permissions": [
        "read:documents",
        "write:documents"
    ]
}
```

Note the role `id` from the response.

## 👥 Assigning Roles to Users

After creating a role, assign it to users or groups.

### Step 1: Create a User Schema

Before creating users, you must define a user schema (user type). This defines the structure and attributes of your users:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/user-schemas \
-d '{
    "name": "test-person",
    "schema": {
        "username": {
            "type": "string"
        },
        "password": {
            "type": "string"
        },
        "email": {
            "type": "string"
        },
        "given_name": {
            "type": "string"
        },
        "family_name": {
            "type": "string"
        }
    }
}'
```

**Response:**

```json
{
    "id": "schema-123",
    "name": "test-person",
    "schema": {
        "username": {"type": "string"},
        "password": {"type": "string"},
        "email": {"type": "string"},
        "given_name": {"type": "string"},
        "family_name": {"type": "string"}
    }
}
```

> **Note:** User schema creation is mandatory. You only need to create a schema once, and it can be reused for all users of that type.

### Step 2: Create a User

Create a user in your organization unit using the schema you just defined:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/users \
-d '{
    "organizationUnit": "<organization-unit-id>",
    "type": "test-person",
    "attributes": {
        "username": "alice",
        "password": "<password>",
        "email": "alice@example.com",
        "given_name": "Alice",
        "family_name": "Smith"
    }
}'
```

Note the user `id` from the response.

### Step 3: Assign Role to User

Assign the role to the user:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/add \
-d '{
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        }
    ]
}'
```

**Alternative: Assign During Role Creation**

You can also assign users when creating the role:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles \
-d '{
    "name": "DocumentEditor",
    "description": "Can read and write documents",
    "ouId": "<organization-unit-id>",
    "permissions": [
        "read:documents",
        "write:documents"
    ],
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        }
    ]
}'
```

## 👥 Assigning Roles to Groups

Roles can also be assigned to groups, making it easier to manage permissions for multiple users.

### Step 1: Create a Group

Create a group in your organization unit:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/groups \
-d '{
    "name": "Editors",
    "description": "Document editors group",
    "organizationUnitId": "<organization-unit-id>"
}'
```

Note the group `id` from the response.

### Step 2: Add Users to Group

Add users to the group:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/groups/<group-id>/users \
-d '{
    "users": ["<user-id-1>", "<user-id-2>"]
}'
```

### Step 3: Assign Role to Group

Assign the role to the group:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/add \
-d '{
    "assignments": [
        {
            "id": "<group-id>",
            "type": "group"
        }
    ]
}'
```

Now all users in the group will inherit the role's permissions.

## 📋 Managing Roles

### List All Roles

Retrieve all roles in an organization unit:

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/roles?ouId=<organization-unit-id>
```

### Get Role Details

Retrieve details of a specific role:

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/roles/<role-id>
```

### Get Role Assignments

List all users and groups assigned to a role:

```bash
curl -kL -H 'Accept: application/json' https://localhost:8090/roles/<role-id>/assignments
```

**Response:**

```json
{
    "assignments": [
        {
            "id": "user-123",
            "type": "user",
            "display": "alice"
        },
        {
            "id": "group-456",
            "type": "group",
            "display": "Editors"
        }
    ]
}
```

### Update Role Permissions

Update the permissions associated with a role:

```bash
curl -kL -X PUT -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id> \
-d '{
    "name": "DocumentEditor",
    "description": "Can read and write documents",
    "ouId": "<organization-unit-id>",
    "permissions": [
        "read:documents",
        "write:documents",
        "delete:documents"
    ]
}'
```

### Remove Role Assignments

Remove users or groups from a role:

```bash
curl -kL -X POST -H 'Content-Type: application/json' https://localhost:8090/roles/<role-id>/assignments/remove \
-d '{
    "assignments": [
        {
            "id": "<user-id>",
            "type": "user"
        }
    ]
}'
```

### Delete a Role

Delete a role (note: all assignments must be removed first):

```bash
curl -kL -X DELETE https://localhost:8090/roles/<role-id>
```

## 🔗 Next Steps

Now that you've set up RBAC, learn how to use it:

- [Flow-Based Authorization](./flow-based-authorization.md) - Integrate authorization into authentication flows
- [OAuth Scopes and Permissions](./oauth-scopes-permissions.md) - Use permissions with OAuth 2.0
