package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Comment struct {
	ID        int       `json:"id"`
	CreatedOn time.Time `json:"created_on"`
	Content   string    `json:"content"`
}

var (
	db        *sql.DB
	clientsMu sync.Mutex
	clients   = make(map[*websocket.Conn]bool)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	maxCommentLength = 500
	maxClients       = 100
)

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func initDB() *sql.DB {
	loadEnv()
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	return db
}

func main() {
	db = initDB()
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(corsMiddleware())
	r.GET("/ws", handleWebSocket)
	r.Use(staticMiddleware())

	log.Println("Server started on http://localhost:8080")
	r.Run(":8080")
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Accept")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func staticMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ws" {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/" || !strings.Contains(c.Request.URL.Path, ".") {
			c.File("./public/index.html")
			c.Abort()
			return
		}

		c.File("./public" + c.Request.URL.Path)
		c.Abort()
	}
}

func handleWebSocket(c *gin.Context) {
	clientsMu.Lock()
	if len(clients) >= maxClients {
		clientsMu.Unlock()
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Max clients reached"})
		return
	}
	clientsMu.Unlock()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer func() {
		conn.Close()
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
	}()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	sendComments(conn)
	handleClientMessages(conn)
}

func handleClientMessages(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var input struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(msg, &input); err != nil {
			continue
		}

		if len(input.Content) > maxCommentLength {
			continue
		}

		saveComment(input.Content)
		broadcastNewComment(input.Content)
	}
}

func sendComments(conn *websocket.Conn) {
	rows, err := db.Query("SELECT id, created_on, content FROM comments ORDER BY created_on ASC LIMIT 50")
	if err != nil {
		log.Printf("Error fetching comments: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.CreatedOn, &comment.Content); err != nil {
			log.Printf("Error scanning comment: %v", err)
			return
		}
		conn.WriteJSON(comment)
	}
}

func saveComment(content string) {
	locJakarta, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(locJakarta)

	_, err := db.Exec("INSERT INTO comments (created_on, content) VALUES (?, ?)", now, content)
	if err != nil {
		log.Printf("Error saving comment: %v", err)
	}
}

func broadcastNewComment(content string) {
	var comment Comment
	err := db.QueryRow("SELECT id, created_on, content FROM comments WHERE content = ? ORDER BY created_on DESC LIMIT 1", content).Scan(&comment.ID, &comment.CreatedOn, &comment.Content)
	if err != nil {
		log.Printf("Error retrieving broadcast comment: %v", err)
		return
	}

	messageJSON, err := json.Marshal(comment)
	if err != nil {
		log.Printf("Error marshaling comment: %v", err)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
