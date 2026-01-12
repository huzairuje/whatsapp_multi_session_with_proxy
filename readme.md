## Overview
---
This project is for engine whatsapp multi-session using library https://github.com/tulir/whatsmeow to use emulate whatsapp web.

### prerequisite
a. gcc (dev essential libs on linux) or using mingw on windows platform
b. golang version >= 1.22

### build up
a. windows platform 
 1. install mingw or gcc from trusted sources like choco or another package manager
 2. build on windows on this command
    ```shell
        set GOOS=windows
        set GOARCH=amd64
        set CGO_ENABLED=1
        go build -o bin/whatsapp_multi_session-windows-amd64.exe
    ```
b. linux or unix platform
 1. install gcc or using `sudo apt install build-essential` 
 2. build on linux to target linux based on your server
    ```shell
        env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-linux-amd64
    ```
    or linux to windows platform
    ```shell
        env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-w -s" -trimpath -o bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
    ```
 3. macOS (using mingw as gcc)
    build on macOS to windows platform
    ```shell
        env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-w -s" -trimpath bin/whatsapp_multi_session_with_proxies-windows-amd64.exe
    ```
### running up
you can run the source code like
```shell
cd PROJECT_ROOT_FOLDER
```
```shell
cp config.local.yaml.example config.local.yaml
```
```shell
go run main.go
```

a. windows platform, open up a command prompt and just cd to the directory and execute via cmd prompt
```shell
  whatsapp_multi_session_with_proxies-windows-amd64.exe
```

a. linux platform, open up a terminal or tmux and just cd to the directory and execute via terminal
```shell
  ./whatsapp_multi_session_with_proxies-linux-amd64
```

a. unix (freebsd or darwin/macOS), open up a terminal or tmux and just cd to the directory and execute via terminal
```shell
  ./whatsapp_multi_session_with_proxies-freebsd-amd64
```
darwin
```shell
  ./whatsapp_multi_session_with_proxies-darwin-amd64
```

---

## Dashboard API Endpoints

This section details the API endpoints available for the dashboard feature.

### Authentication

#### POST /dashboard/auth/register
*   **Description:** Registers a new admin user for the dashboard. The phone number provided will be used for OTP verification during login.
*   **Authentication Required:** No
*   **Request Body:**
    *   `phone_number` (string, required): The WhatsApp phone number of the admin user (e.g., "6281234567890").
*   **Example Request:**
    ```json
    {
        "phone_number": "6281234567890"
    }
    ```
*   **Example Successful Response (201 Created):**
    ```json
    {
        "message": "Admin user registered successfully. Please request an OTP to login."
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input (e.g., phone number missing or invalid format).
    *   `409 Conflict`: Phone number already registered.
    *   `500 Internal Server Error`: Server-side issue.

#### POST /dashboard/auth/request-otp
*   **Description:** Requests an OTP (One-Time Password) to be sent to the registered admin user's WhatsApp number. The OTP is required for login.
*   **Authentication Required:** No
*   **Request Body:**
    *   `phone_number` (string, required): The registered WhatsApp phone number of the admin user.
*   **Example Request:**
    ```json
    {
        "phone_number": "6281234567890"
    }
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "message": "OTP sent successfully to 6281234567890"
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input.
    *   `404 Not Found`: Admin user with the given phone number not found.
    *   `500 Internal Server Error`: Failed to generate OTP, save OTP, or send OTP message (e.g., OTP sending service not configured or WhatsApp message sending failed).

#### POST /dashboard/auth/login
*   **Description:** Logs in an admin user using their phone number and the OTP they received.
*   **Authentication Required:** No
*   **Request Body:**
    *   `phone_number` (string, required): The registered WhatsApp phone number of the admin user.
    *   `otp` (string, required): The OTP received by the admin user.
*   **Example Request:**
    ```json
    {
        "phone_number": "6281234567890",
        "otp": "123456"
    }
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input.
    *   `401 Unauthorized`: Invalid OTP, OTP expired, or OTP not requested/already used.
    *   `404 Not Found`: Admin user not found.
    *   `500 Internal Server Error`: Failed to generate JWT or database error.

#### GET /dashboard/auth/detail
*   **Description:** Retrieves the details of the currently authenticated admin user.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:** None
*   **Example Request:**
    ```
    GET /dashboard/auth/detail
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "id": 1,
        "phone_number": "6281234567890",
        "created_at": "2023-10-27T10:00:00Z"
    }
    ```
*   **Common Error Responses:**
    *   `401 Unauthorized`: Invalid or missing JWT token, or token expired.
    *   `404 Not Found`: Authenticated admin user not found in the database (should be rare if token is valid).
    *   `500 Internal Server Error`: Server-side issue.

#### POST /dashboard/auth/logout
*   **Description:** Logs out the currently authenticated admin user. For stateless JWT, this typically means the client should discard the token.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:** None
*   **Example Request:**
    ```
    POST /dashboard/auth/logout
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "message": "Logged out successfully. Please discard your token."
    }
    ```
*   **Common Error Responses:**
    *   `401 Unauthorized`: Invalid or missing JWT token.

### Device Management

#### POST /dashboard/devices
*   **Description:** Adds a new WhatsApp device (JID) to be managed by the authenticated admin user. The device must already be known/registered by the underlying WhatsApp system.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:**
    *   `device_jid` (string, required): The JID of the WhatsApp device to manage (e.g., "1122334455@s.whatsapp.net").
    *   `device_name` (string, required): A friendly name for the device.
*   **Example Request:**
    ```json
    {
        "device_jid": "1122334455@s.whatsapp.net",
        "device_name": "Work Device Alpha"
    }
    ```
*   **Example Successful Response (201 Created):**
    ```json
    {
        "id": 1,
        "admin_user_id": 1,
        "device_jid": "1122334455@s.whatsapp.net",
        "device_name": "Work Device Alpha",
        "created_at": "2023-10-27T11:00:00Z",
        "updated_at": "2023-10-27T11:00:00Z"
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input (e.g., JID format, missing fields).
    *   `401 Unauthorized`: Invalid or missing JWT token.
    *   `404 Not Found`: Device JID not registered/recognized by the underlying WhatsApp system.
    *   `409 Conflict`: Device JID already managed by this admin user.
    *   `500 Internal Server Error`: Database error or other server-side issue.

#### GET /dashboard/devices
*   **Description:** Lists all WhatsApp devices managed by the authenticated admin user, including their live login status.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:** None
*   **Example Request:**
    ```
    GET /dashboard/devices
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```
*   **Example Successful Response (200 OK):**
    ```json
    [
        {
            "id": 1,
            "admin_user_id": 1,
            "device_jid": "1122334455@s.whatsapp.net",
            "device_name": "Work Device Alpha",
            "is_logged_in_live": true,
            "created_at": "2023-10-27T11:00:00Z",
            "updated_at": "2023-10-27T11:00:00Z"
        },
        {
            "id": 2,
            "admin_user_id": 1,
            "device_jid": "6677889900@s.whatsapp.net",
            "device_name": "Personal Device Beta",
            "is_logged_in_live": false,
            "created_at": "2023-10-28T12:00:00Z",
            "updated_at": "2023-10-28T12:00:00Z"
        }
    ]
    ```
*   **Common Error Responses:**
    *   `401 Unauthorized`: Invalid or missing JWT token.
    *   `500 Internal Server Error`: Database error or error fetching live status.

### Message Sending & Reporting

#### POST /dashboard/messages
*   **Description:** Sends a single WhatsApp message from a managed device to a specified recipient.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:**
    *   `device_jid` (string, required): The JID of the managed device to send from.
    *   `recipient_number` (string, required): The recipient's phone number (e.g., "6289876543210").
    *   `message_content` (string, required): The text content of the message.
*   **Example Request:**
    ```json
    {
        "device_jid": "1122334455@s.whatsapp.net",
        "recipient_number": "6289876543210",
        "message_content": "Hello from the dashboard!"
    }
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "message": "Message sent successfully",
        "message_id": 101, // Database ID of the sent message record
        "whatsapp_message_id": "1234567890ABCDEF", // WhatsApp's message ID
        "status": "sent"
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input.
    *   `401 Unauthorized`: Invalid or missing JWT token.
    *   `403 Forbidden`: Selected `device_jid` is not managed by the admin user.
    *   `412 Precondition Failed`: Selected `device_jid` is not currently logged in.
    *   `500 Internal Server Error`: Failed to send message via WhatsApp or failed to record message in DB.

#### POST /dashboard/messages/bulk
*   **Description:** Sends the same WhatsApp message from a managed device to multiple recipients.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Request Body:**
    *   `device_jid` (string, required): The JID of the managed device to send from.
    *   `recipient_numbers` (array of strings, required): A list of recipient phone numbers.
    *   `message_content` (string, required): The text content of the message.
*   **Example Request:**
    ```json
    {
        "device_jid": "1122334455@s.whatsapp.net",
        "recipient_numbers": ["6289876543210", "6281122334455"],
        "message_content": "This is a bulk message."
    }
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "overall_status": "completed_with_errors", // or "completed_successfully"
        "results": [
            {
                "recipient_number": "6289876543210",
                "status": "sent",
                "message_id": "1234567890ABCXYZ" // WhatsApp's message ID
            },
            {
                "recipient_number": "6281122334455",
                "status": "failed",
                "error": "Failed to send message: Recipient not found or blocked"
            }
        ]
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid input.
    *   `401 Unauthorized`: Invalid or missing JWT token.
    *   `403 Forbidden`: Selected `device_jid` is not managed by the admin user.
    *   `412 Precondition Failed`: Selected `device_jid` is not currently logged in.
    *   `500 Internal Server Error`: Problem with the messaging service or database recording (for individual messages, this might be reflected in the `results`).

#### GET /dashboard/report/messages
*   **Description:** Retrieves a paginated report of messages sent by the authenticated admin user.
*   **Authentication Required:** Yes (JWT Bearer Token)
*   **Query Parameters:**
    *   `page` (integer, optional, default: 1): The page number for pagination.
    *   `page_size` (integer, optional, default: 10): The number of messages per page (max: 100).
*   **Example Request:**
    ```
    GET /dashboard/report/messages?page=1&page_size=20
    Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
    ```
*   **Example Successful Response (200 OK):**
    ```json
    {
        "messages": [
            {
                "id": 101,
                "admin_user_id": 1,
                "device_jid": "1122334455@s.whatsapp.net",
                "recipient_number": "6289876543210",
                "message_content": "Hello from the dashboard!",
                "status": "sent",
                "sent_at": "2023-10-27T12:00:00Z",
                "message_id_from_whatsapp": "1234567890ABCDEF",
                "created_at": "2023-10-27T12:00:05Z"
            }
            // ... more messages
        ],
        "page": 1,
        "page_size": 20,
        "total_count": 50,
        "total_pages": 3
    }
    ```
*   **Common Error Responses:**
    *   `400 Bad Request`: Invalid pagination parameters.
    *   `401 Unauthorized`: Invalid or missing JWT token.
    *   `500 Internal Server Error`: Database error.
