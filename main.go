package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
}


var db *sql.DB

func main() {
	initDB()
	defer db.Close()

	// Create a new Gin router
	router := gin.Default()

	// Define API endpoints
	router.POST("/tasks", createTask)
	router.GET("/tasks/:id", getTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)
	router.GET("/tasks", listTasks)

	// Start the server
	router.Run()
}


func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Create the tasks table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT ,
		description TEXT,
		due_date TEXT,
		status TEXT
	);`)
	if err != nil {
		log.Fatal("Failed to create the tasks table:", err)
	}
}


func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	result, err := db.Exec(`INSERT INTO tasks (title, description, due_date, status)
		VALUES (?, ?, ?, ?);`, task.Title, task.Description, task.DueDate, task.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create the task"})
		return
	}

	// Get the ID of the created task
	id, _ := result.LastInsertId()

	task.ID = int(id)
	c.JSON(http.StatusOK, task)
}

// Retrieve a task
func getTask(c *gin.Context) {
	id := c.Param("id")

	var task Task
	err := db.QueryRow(`SELECT id, title, description, due_date, status FROM tasks WHERE id = ?;`, id).
		Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}


func updateTask(c *gin.Context) {
	id := c.Param("id")

	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the task in the database
	_, err := db.Exec(`UPDATE tasks SET title = ?, description = ?, due_date = ?, status = ? WHERE id = ?;`,
		task.Title, task.Description, task.DueDate, task.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the task"})
		return
	}

	task.ID, _ = strconv.Atoi(id)
	c.JSON(http.StatusOK, task)
}

// Delete a task
func deleteTask(c *gin.Context) {
	id := c.Param("id")

	// Delete the task from the database
	_, err := db.Exec(`DELETE FROM tasks WHERE id = ?;`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// List all tasks
func listTasks(c *gin.Context) {
	rows, err := db.Query(`SELECT id, title, description, due_date, status FROM tasks;`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate, &task.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
			return
		}
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}
