# Go-Chat

[![Go version](https://img.shields.io/badge/go-1.21-blue)](https://golang.org)

A simple real-time chat application built with Go and React, designed as a learning project to deepen understanding of Golang backend development and modern frontend integration.

## Overview
Go-Chat is a real-time chat system implemented with a Go backend and a React frontend. The project focuses on practicing Go language skills, including REST APIs, WebSocket communication, token-based authentication, and state management.

---

## Features
- User registration and login with token-based authentication
- JWT access and refresh tokens handling with token refresh flow
- Real-time chat using WebSockets
- Clean React frontend with TypeScript and functional components
- Basic UI with login, register, lobby, chat, and token refresh pages
- Proper error handling and loading states in the frontend

---

## Technologies Used
- **Backend:** Go (Golang) with protobuf message encoding
- **Frontend:** React with TypeScript
- **Communication:** REST API and WebSocket for real-time messaging
- **Authentication:** JWT access and refresh tokens
- **State Management:** React hooks and context

---

## Installation
### Backend
1. Install Go 1.21 or later.
2. Clone the repository:
    ```bash
    git clone https://github.com/FilipeJohansson/go-chat.git
    cd go-chat/backend
    ```
3. Run the backend server:
    ```bash
    go run main.go
    ```

### Frontend
1. Navigate to the frontend directory:
    ```bash
    cd go-chat/frontend
    ```
2. Install dependencies:
    ```bash
    npm install
    ```
3. Run the development server:
    ```bash
    npm start
    ```
4. Open http://localhost:5174 in your browser.
---

## Usage
- Register a new user via the Register page.
- Login using your credentials.
- Tokens are managed automatically, including refresh tokens.
- After login, you can enter the chat lobby and start real-time conversations.

---

## Docker
This project includes Docker support to simplify setup and deployment for both backend and frontend.

### Using Docker
#### Build and Run with Docker Compose

1. Make sure you have [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/) installed.
2. From the root of the project, run:
    ```bash
    docker-compose up --build
    ```
    This command will:
    - Build the Go backend using the provided `Dockerfile` in the `backend/` directory
    - Build the React frontend using the `Dockerfile` in the `frontend/` directory
    - Start both services and expose them on the following ports:
      - Frontend: http://localhost:5174
      - Backend: http://localhost:8080

3. Access the frontend at http://localhost:5174
