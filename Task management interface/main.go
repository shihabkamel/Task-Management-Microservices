package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

// Task struct for task data
type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

var db *sql.DB

func main() {
	// Initialize SQLite database
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create tasks table if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status TEXT NOT NULL
	)`)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	// Create a task
	r.POST("/tasks", func(c *gin.Context) {
		var newTask Task
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newTask.Status = "pending"

		result, err := db.Exec("INSERT INTO tasks (title, status) VALUES (?, ?)", newTask.Title, newTask.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, _ := result.LastInsertId()
		newTask.ID = int(id)
		c.JSON(http.StatusCreated, newTask)
	})

	// Get all tasks
	r.GET("/tasks", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, title, status FROM tasks")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var task Task
			if err := rows.Scan(&task.ID, &task.Title, &task.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			tasks = append(tasks, task)
		}
		c.JSON(http.StatusOK, tasks)
	})

	// Update a task
	r.PUT("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		var updatedTask Task
		if err := c.ShouldBindJSON(&updatedTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec("UPDATE tasks SET title = ?, status = ? WHERE id = ?", updatedTask.Title, updatedTask.Status, idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		updatedTask.ID = idInt
		c.JSON(http.StatusOK, updatedTask)
	})

	// Delete a task
	r.DELETE("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		result, err := db.Exec("DELETE FROM tasks WHERE id = ?", idInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
	})

	// Start the server
	r.Run(":8080")
}
