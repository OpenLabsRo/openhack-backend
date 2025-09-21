# API Documentation

This document outlines the API endpoints for the OpenHack backend application.

## Base URL

The base URL for all endpoints is `/`.

## Authentication

Most endpoints require authentication via a JWT token provided in the `Authorization` header as a Bearer token.

---

## General

### GET /ping

- **Description:** A health check endpoint.
- **Success Response (200):** `PONG`

### GET /version

- **Description:** Returns the current version of the backend.
- **Success Response (200):** The version string.

---

## Accounts

### POST /accounts/check

- **Description:** Checks if an account with the given email is registered.
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "registered": true
  }
  ```
- **Error Responses:**
  - `404 Not Found`: Account not found.

### POST /accounts/register

- **Description:** Registers a new account.
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "token": "your.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `400 Bad Request`: Account already registered.
  - `500 Internal Server Error`: Internal server error.

### POST /accounts/login

- **Description:** Logs in an existing account.
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "token": "your.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Wrong password.
  - `404 Not Found`: Account not found.

### GET /accounts/whoami

- **Description:** Returns the currently authenticated account.
- **Success Response (200):**
  ```json
  {
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.

### PATCH /accounts

- **Description:** Edits the name of the currently authenticated account.
- **Request Body:**
  ```json
  {
    "name": "New Name"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "token": "your.new.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

---

## Teams

### GET /teams

- **Description:** Gets the team of the currently authenticated account.
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.

### POST /teams

- **Description:** Creates a new team for the currently authenticated account.
- **Request Body:**
  ```json
  {
    "name": "My Awesome Team"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "token": "your.new.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `400 Bad Request`: Account already has a team.
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams

- **Description:** Changes the name of the team for the currently authenticated account.
- **Request Body:**
  ```json
  {
    "name": "New Team Name"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### DELETE /teams

- **Description:** Deletes the team of the currently authenticated account. Only works if the team has only one member.
- **Success Response (200):**
  ```json
  {
    "token": "your.new.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `400 Bad Request`: Team is not empty.
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### GET /teams/members

- **Description:** Gets the members of the team for the currently authenticated account.
- **Success Response (200):**
  ```json
  [
    { "account": { ... } },
    { "account": { ... } }
  ]
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/join

- **Description:** Joins a team.
- **Query Parameters:**
  - `id`: The ID of the team to join.
- **Success Response (200):**
  ```json
  {
    "token": "your.new.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `400 Bad Request`: Account already has a team.
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Team not found.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/leave

- **Description:** Leaves the current team.
- **Success Response (200):**
  ```json
  {
    "token": "your.new.jwt.token",
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team or team not found.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/kick

- **Description:** Kicks a member from the team.
- **Query Parameters:**
  - `id`: The ID of the account to kick.
- **Success Response (200):** `OK`
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account not found or team not found.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/submissions/name

- **Description:** Changes the submission name for the team.
- **Request Body:**
  ```json
  {
    "name": "New Submission Name"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/submissions/desc

- **Description:** Changes the submission description for the team.
- **Request Body:**
  ```json
  {
    "desc": "New Submission Description"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/submissions/repo

- **Description:** Changes the submission repository URL for the team.
- **Request Body:**
  ```json
  {
    "repo": "https://github.com/user/repo"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

### PATCH /teams/submissions/pres

- **Description:** Changes the submission presentation URL for the team.
- **Request Body:**
  ```json
  {
    "pres": "https://example.com/presentation"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "team": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account has no team.
  - `500 Internal Server Error`: Internal server error.

---

## Superusers

This section is for admin-level users.

### POST /superusers/login

- **Description:** Logs in a superuser.
- **Request Body:**
  ```json
  {
    "username": "admin",
    "password": "password123"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "token": "your.jwt.token",
    "superuser": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Wrong password.
  - `404 Not Found`: Superuser not found.

### GET /superusers/whoami

- **Description:** Returns the currently authenticated superuser.
- **Success Response (200):**
  ```json
  {
    "superuser": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.

### POST /superusers/accounts/initialize

- **Description:** Initializes a new account.
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "account": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### GET /superusers/flags

- **Description:** Gets all feature flags.
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### POST /superusers/flags

- **Description:** Sets a feature flag.
- **Request Body:**
  ```json
  {
    "flag": "flag_name",
    "value": true
  }
  ```
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### PUT /superusers/flags

- **Description:** Sets multiple feature flags.
- **Request Body:**
  ```json
  {
    "flag1": true,
    "flag2": false
  }
  ```
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### DELETE /superusers/flags

- **Description:** Unsets a feature flag.
- **Request Body:**
  ```json
  {
    "flag": "flag_name"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### PUT /superusers/flags/reset

- **Description:** Resets all feature flags to their default values.
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### GET /superusers/flags/stages

- **Description:** Gets all flag stages.
- **Success Response (200):**
  ```json
  [
    { "stage": { ... } },
    { "stage": { ... } }
  ]
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### POST /superusers/flags/stages

- **Description:** Creates a new flag stage.
- **Request Body:**
  ```json
  {
    "name": "Stage Name",
    "flags": { ... }
  }
  ```
- **Success Response (200):**
  ```json
  {
    "stage": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### DELETE /superusers/flags/stages

- **Description:** Deletes a flag stage.
- **Query Parameters:**
  - `id`: The ID of the flag stage to delete.
- **Success Response (200):**
  ```json
  {
    "stage": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### POST /superusers/flags/stages/execute

- **Description:** Executes a flag stage.
- **Query Parameters:**
  - `id`: The ID of the flag stage to execute.
- **Success Response (200):**
  ```json
  {
    "flags": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Flag stage not found.
  - `500 Internal Server Error`: Internal server error.

### GET /superusers/checkin/badges

- **Description:** Gets all accounts grouped by badge piles.
- **Success Response (200):**
  ```json
  [
    [ { "account": { ... } }, ... ],
    [ { "account": { ... } }, ... ]
  ]
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### GET /superusers/checkin/tags

- **Description:** Gets a tag by ID.
- **Query Parameters:**
  - `id`: The ID of the tag.
- **Success Response (200):**
  ```json
  {
    "tag": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Tag not found.

### POST /superusers/checkin/tags

- **Description:** Assigns a tag to an account.
- **Request Body:**
  ```json
  {
    "id": "tag_id",
    "account_id": "account_id"
  }
  ```
- **Success Response (200):**
  ```json
  {
    "tag": { ... }
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `500 Internal Server Error`: Internal server error.

### POST /superusers/checkin/scan

- **Description:** Scans a participant's QR code for check-in.
- **Query Parameters:**
  - `id`: The ID of the account to check in.
- **Success Response (200):**
  ```json
  {
    "account": { ... },
    "pile": 1
  }
  ```
- **Error Responses:**
  - `401 Unauthorized`: Invalid or missing token.
  - `404 Not Found`: Account not found.
