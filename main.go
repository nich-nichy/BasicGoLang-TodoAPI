package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Task struct {
	ID        int    `json:"id"`
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
}

var (
	tasks   []Task
	taskID  int
	taskMux sync.Mutex
)

func main() {
	http.HandleFunc("/tasks", handleTasks)     // GET, POST
	http.HandleFunc("/tasks/", handleTaskByID) // DELETE, PATCH
	http.HandleFunc("/", welcomeMessage)       // Default message

	fmt.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func welcomeMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the Go TODO API!")
	fmt.Fprintln(w, "Use /tasks endpoint to manage your TODOs.")
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasks(w)
	case http.MethodPost:
		addTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractTaskID(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		deleteTask(w, id)
	case http.MethodPatch:
		markTaskCompleted(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTasks(w http.ResponseWriter) {
	taskMux.Lock()
	defer taskMux.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func addTask(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, "Invalid task data", http.StatusBadRequest)
		return
	}

	taskMux.Lock()
	defer taskMux.Unlock()

	taskID++
	newTask.ID = taskID
	newTask.Completed = false
	tasks = append(tasks, newTask)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func deleteTask(w http.ResponseWriter, id int) {
	taskMux.Lock()
	defer taskMux.Unlock()

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func markTaskCompleted(w http.ResponseWriter, id int) {
	taskMux.Lock()
	defer taskMux.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Completed = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func extractTaskID(path string) (int, error) {
	var id int
	_, err := fmt.Sscanf(path, "/tasks/%d", &id)
	return id, err
}
