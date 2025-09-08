-- Table to store Organization Units
CREATE TABLE ORGANIZATION_UNIT (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    OU_ID       VARCHAR(36) UNIQUE NOT NULL,
    PARENT_ID   VARCHAR(36),
    HANDLE      VARCHAR(50)        NOT NULL,
    NAME        VARCHAR(50)        NOT NULL,
    DESCRIPTION VARCHAR(255),
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (PARENT_ID) REFERENCES ORGANIZATION_UNIT (OU_ID) ON DELETE CASCADE
);

-- Table to store Users
CREATE TABLE USER (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    USER_ID     VARCHAR(36) UNIQUE NOT NULL,
    OU_ID       VARCHAR(36)        NOT NULL,
    TYPE        TEXT               NOT NULL,
    ATTRIBUTES  TEXT,
    CREDENTIALS TEXT,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now'))
);

-- Table to store Groups
CREATE TABLE "GROUP" (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    GROUP_ID    VARCHAR(36) UNIQUE NOT NULL,
    OU_ID       VARCHAR(36)        NOT NULL,
    NAME        VARCHAR(50)        NOT NULL,
    DESCRIPTION VARCHAR(255),
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now'))
);

-- Table to store Group member assignments
CREATE TABLE GROUP_MEMBER_REFERENCE (
    ID          INTEGER PRIMARY KEY AUTOINCREMENT,
    GROUP_ID    VARCHAR(36) NOT NULL,
    MEMBER_TYPE VARCHAR(7)  NOT NULL,
    MEMBER_ID   VARCHAR(36) NOT NULL,
    CREATED_AT  TEXT DEFAULT (datetime('now')),
    UPDATED_AT  TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (GROUP_ID) REFERENCES "GROUP" (GROUP_ID) ON DELETE CASCADE
);

-- Table to store basic service provider (app) details.
CREATE TABLE SP_APP (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    APP_ID VARCHAR(36) UNIQUE NOT NULL,
    APP_NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(50) NOT NULL,
    AUTH_FLOW_GRAPH_ID VARCHAR(50) NOT NULL,
    REGISTRATION_FLOW_GRAPH_ID VARCHAR(50) NOT NULL,
    IS_REGISTRATION_FLOW_ENABLED CHAR(1) DEFAULT '1',
    APP_JSON TEXT
);

-- Table to store OAuth configurations for SP apps.
CREATE TABLE IDN_OAUTH_CONSUMER_APPS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CONSUMER_KEY VARCHAR(255) NOT NULL,
    CONSUMER_SECRET VARCHAR(255) NOT NULL,
    APP_ID VARCHAR(36) NOT NULL,
    OAUTH_CONFIG_JSON TEXT,
    FOREIGN KEY (APP_ID) REFERENCES SP_APP(APP_ID) ON DELETE CASCADE
);

-- Table to store inbound auth configs (e.g., OAuth, SAML) for SP apps.
CREATE TABLE SP_INBOUND_AUTH (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    INBOUND_AUTH_KEY VARCHAR(255) NOT NULL,
    INBOUND_AUTH_TYPE VARCHAR(50) NOT NULL,
    APP_ID VARCHAR(36) NOT NULL,
    FOREIGN KEY (APP_ID) REFERENCES SP_APP(APP_ID) ON DELETE CASCADE
);

-- Table to store identity providers.
CREATE TABLE IDP (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    IDP_ID VARCHAR(36) UNIQUE NOT NULL,
    NAME VARCHAR(255) NOT NULL,
    DESCRIPTION VARCHAR(500),
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Table to store identity provider properties.
CREATE TABLE IDP_PROPERTY (
    IDP_ID VARCHAR(36) NOT NULL,
    PROPERTY_NAME VARCHAR(255) NOT NULL,
    PROPERTY_VALUE VARCHAR(500) NOT NULL,
    IS_SECRET CHAR(1) DEFAULT '0',
    PRIMARY KEY (IDP_ID, PROPERTY_NAME),
    FOREIGN KEY (IDP_ID) REFERENCES IDP(IDP_ID) ON DELETE CASCADE
);

-- Table to store notification senders.
CREATE TABLE NOTIFICATION_SENDER (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    NAME VARCHAR(255) NOT NULL,
    SENDER_ID VARCHAR(36) UNIQUE NOT NULL,
    DESCRIPTION VARCHAR(500),
    TYPE VARCHAR(20) NOT NULL,
    PROVIDER VARCHAR(20) NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Table to store notification sender properties.
CREATE TABLE NOTIFICATION_SENDER_PROPERTY (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    SENDER_ID INTEGER NOT NULL,
    PROPERTY_NAME VARCHAR(255) NOT NULL,
    PROPERTY_VALUE VARCHAR(500),
    IS_SECRET CHAR(1) DEFAULT '0',
    UNIQUE (SENDER_ID, PROPERTY_NAME),
    FOREIGN KEY (SENDER_ID) REFERENCES NOTIFICATION_SENDER(SENDER_ID) ON DELETE CASCADE
);

-- Table to store certificates associated with various entities.
CREATE TABLE CERTIFICATE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CERT_ID VARCHAR(36) UNIQUE NOT NULL,
    REF_TYPE VARCHAR(20) NOT NULL,
    REF_ID VARCHAR(36) NOT NULL,
    TYPE VARCHAR(20) NOT NULL,
    VALUE TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now')),
    UNIQUE (REF_TYPE, REF_ID)
);

-- Table to store user schemas.
CREATE TABLE USER_SCHEMAS (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    SCHEMA_ID VARCHAR(36) UNIQUE NOT NULL,
    NAME VARCHAR(255) UNIQUE NOT NULL,
    SCHEMA_DEF TEXT NOT NULL,
    CREATED_AT TEXT DEFAULT (datetime('now')),
    UPDATED_AT TEXT DEFAULT (datetime('now'))
);

-- Insert sample data into the tables.
INSERT INTO SP_APP (APP_NAME, APP_ID, DESCRIPTION, AUTH_FLOW_GRAPH_ID, REGISTRATION_FLOW_GRAPH_ID, APP_JSON) 
VALUES ('Test SPA', '550e8400-e29b-41d4-a716-446655440000', 'Initial testing App', 'auth_flow_config_basic', 'registration_flow_config_basic',
'{"url": "https://localhost:3000", "logo_url": "https://localhost:3000/logo.png", "token": {"issuer": "thunder", "validity_period": 3600, "user_attributes": ["email", "username"]}}');

INSERT INTO IDN_OAUTH_CONSUMER_APPS (CONSUMER_KEY, CONSUMER_SECRET, APP_ID, OAUTH_CONFIG_JSON)
VALUES ('client123', 'fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4', '550e8400-e29b-41d4-a716-446655440000', 
'{"redirect_uris":["https://localhost:3000"],"grant_types":["client_credentials","authorization_code","refresh_token"],"response_types":["code"],"token_endpoint_auth_methods":["client_secret_basic","client_secret_post"],"token":{"access_token":{"issuer":"thunder","validity_period":3600,"user_attributes":["email","username"]}}}');

INSERT INTO SP_INBOUND_AUTH (INBOUND_AUTH_KEY, INBOUND_AUTH_TYPE, APP_ID)
VALUES ('client123', 'oauth2', '550e8400-e29b-41d4-a716-446655440000');

INSERT INTO USER (USER_ID, OU_ID, TYPE, ATTRIBUTES, CREATED_AT, UPDATED_AT)
VALUES (
           '550e8400-e29b-41d4-a716-446655440000', '456e8400-e29b-41d4-a716-446655440001', 'person',
           '{"age": 30, "roles": ["admin", "user"], "address": {"city": "Colombo", "zip": "00100"}}',
           datetime('now'), datetime('now')
       );

INSERT INTO IDP (IDP_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT)
VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Local', 'Local Identity Provider', datetime('now'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440001', 'Github', 'Login with Github', datetime('now'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440002', 'Google', 'Login with Google', datetime('now'), datetime('now'));

INSERT INTO IDP_PROPERTY (IDP_ID, PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET)
VALUES
('550e8400-e29b-41d4-a716-446655440001', 'client_id', 'client1', '0'),
('550e8400-e29b-41d4-a716-446655440001', 'client_secret', 'secret1', '1'),
('550e8400-e29b-41d4-a716-446655440001', 'redirect_uri', 'https://localhost:3000', '0'),
('550e8400-e29b-41d4-a716-446655440001', 'scopes', 'user:email,read:user', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'client_id', 'client2', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'client_secret', 'secret2', '1'),
('550e8400-e29b-41d4-a716-446655440002', 'redirect_uri', 'https://localhost:3000', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'scopes', 'openid,email,profile', '0');

-- Insert sample organization units
INSERT INTO ORGANIZATION_UNIT (OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT)
VALUES
('456e8400-e29b-41d4-a716-446655440001', NULL, 'root', 'Root Organization', 'Root organization unit', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440002', '456e8400-e29b-41d4-a716-446655440001', 'engineering', 'Engineering', 'Engineering department', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440003', '456e8400-e29b-41d4-a716-446655440001', 'sales', 'Sales', 'Sales department', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440004', '456e8400-e29b-41d4-a716-446655440002', 'frontend', 'Frontend Team', 'Frontend development team', datetime('now'), datetime('now'));

-- Insert sample user schemas
INSERT INTO USER_SCHEMAS (SCHEMA_ID, NAME, SCHEMA_DEF, CREATED_AT, UPDATED_AT)
VALUES
('550e8400-e29b-41d4-a716-446655440100', 'basic-user-schema', '{"type": "object", "properties": {"firstName": {"type": "string"}, "lastName": {"type": "string"}, "email": {"type": "string", "regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$"}}, "required": ["firstName", "lastName", "email"]}', datetime('now'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440101', 'employee-profile-schema', '{"type": "object", "properties": {"personalInfo": {"type": "object", "properties": {"firstName": {"type": "string"}, "lastName": {"type": "string"}, "dateOfBirth": {"type": "string", "regex": "^\\\\d{4}-\\\\d{2}-\\\\d{2}$"}}}, "workInfo": {"type": "object", "properties": {"department": {"type": "string"}, "position": {"type": "string"}, "startDate": {"type": "string", "regex": "^\\\\d{4}-\\\\d{2}-\\\\d{2}$"}}}, "contact": {"type": "object", "properties": {"email": {"type": "string", "regex": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$"}, "phone": {"type": "string"}}}, "credentials": {"type": "object", "properties": {"password": {"type": "string", "regex": "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\\\\\\\d)(?=.*[!@#$%^&*()_+\\\\\\\\-=\\\\\\\\[\\\\\\\\]{};'':\"\\\\\\\\\\\\\\\\|,.<>\\\\\\\\/?]).{8,}$"}}}}, "required": ["personalInfo", "workInfo", "contact", "credentials"]}', datetime('now'), datetime('now'));
