# Comment App

A simple Go-based web application for handling comments with WebSocket support.  

## Features
- RESTful API for managing comments.
- Real-time updates via WebSocket.
- Environment configuration with `.env` support.

## Technologies Used
- **Gin**: HTTP web framework ([github.com/gin-gonic/gin](https://github.com/gin-gonic/gin))
- **Go-MySQL-Driver**: MySQL database driver ([github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))
- **Gorilla WebSocket**: WebSocket protocol implementation ([github.com/gorilla/websocket](https://github.com/gorilla/websocket))
- **Godotenv**: .env file parser ([github.com/joho/godotenv](https://github.com/joho/godotenv))

## Requirements
- Go 1.20 or higher installed.
- MySQL database setup with the required schema.

## Getting Started

1. **Clone the repository**:
   ```bash
   git clone https://github.com/jtnqr/comment_app.git
   cd comment_app
   ```

2. **Set up your .env file**: 
    Copy .env.example to .env and configure your database connection settings:

    ```env
    DB_USER=your_db_user
    DB_PASS=your_db_password
    DB_NAME=your_db_name
    DB_HOST=your_db_host
    DB_PORT=3306
    ```
    Ensure your MySQL database is running: Create the required database and tables if they do not already exist.

3. **Run the application**:

    ```bash
    go run main.go
    ```
4. **Access the app**: 
    Open your browser or API client and navigate to the base URL:
    http://localhost:8080

## Contributing
1. **Fork the repository.**
2. **Create your feature branch**:
    ```bash
    git checkout -b my-new-feature
    ```
3. **Commit your changes**:
    ```bash
    git commit -am 'Add some feature'
    ```
4. **Push to the branch**:
    ```bash
    git push origin my-new-feature
    ```
5. **Submit a pull request**.
