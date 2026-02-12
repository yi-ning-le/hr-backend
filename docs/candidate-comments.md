# Candidate Comments API Documentation

This document describes the API endpoints for the candidate commenting system (Twitter-style thread).

## Base URL

All endpoints are relative to: `import.meta.env.VITE_API_BASE_URL` (Default: `/api`)

## Data Models

### CandidateComment

| Field          | Type                    | Description                                              |
| :------------- | :---------------------- | :------------------------------------------------------- |
| `id`           | `string`                | Unique identifier for the comment                        |
| `candidateId`  | `string`                | ID of the candidate this comment belongs to              |
| `authorId`     | `string`                | ID of the employee who wrote the comment                 |
| `authorName`   | `string`                | Full name of the author                                  |
| `authorAvatar` | `string` (Optional)     | URL to the author's avatar image                         |
| `authorRole`   | `"HR" \| "INTERVIEWER"` | The functional role of the author at the time of posting |
| `content`      | `string`                | The markdown or plain text content of the comment        |
| `createdAt`    | `string` (ISO 8601)     | Timestamp of when the comment was created                |

---

## Endpoints

### 1. List Candidate Comments

Retrieves all comments for a specific candidate, ordered by creation date (newest first).

- **URL:** `/candidates/:candidateId/comments`
- **Method:** `GET`
- **Auth Required:** Yes
- **Success Response:**
  - **Code:** 200 OK
  - **Content:** `CandidateComment[]`

### 2. Create Candidate Comment

Posts a new comment for a specific candidate.

- **URL:** `/candidates/:candidateId/comments`
- **Method:** `POST`
- **Auth Required:** Yes
- **Data Params:**
  ```json
  {
    "content": "string"
  }
  ```
- **Success Response:**
  - **Code:** 201 Created
  - **Content:** `CandidateComment`

### 3. Delete Comment

Deletes a specific comment by its ID.

- **URL:** `/comments/:commentId`
- **Method:** `DELETE`
- **Auth Required:** Yes
- **Success Response:**
  - **Code:** 204 No Content

---

## Implementation Notes

- **Author Identity:** The backend should derive `authorId`, `authorName`, and `authorRole` from the authenticated session/JWT token.
- **Ordering:** Comments should ideally be returned in descending order of `createdAt` to match the Twitter-style UI requirements.
- **Permissions:**
  - Only HR and assigned Interviewers should be able to post/view comments.
  - Users should only be able to delete their own comments (or Admin/HR can delete any if required by policy).
