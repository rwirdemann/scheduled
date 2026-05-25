package file

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rwirdemann/scheduled"
)

// TaskRepository stores tasks in a JSON file.
type TaskRepository struct {
	path string
}

// NewTaskRepository creates a new TaskRepository. taskPath must be the full
// path to a .json file. The directory is created if it does not exist.
func NewTaskRepository(path string) TaskRepository {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}
	return TaskRepository{path: path}
}

// LoadTasks loads and returns all tasks from the repository file.
func (t TaskRepository) LoadTasks() []scheduled.Task {
	file, err := os.Open(t.path)
	if err != nil {
		return []scheduled.Task{}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var tasks struct {
		Tasks []scheduled.Task `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		log.Printf("Failed to decode %s: %v", t.path, err)
		return []scheduled.Task{}
	}

	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == "" {
			tasks.Tasks[i].ID = uuid.NewString()
		}
	}

	return tasks.Tasks
}

// Upsert upserts a single task in the repository file.
func (t TaskRepository) Upsert(task scheduled.Task) {
	tasks := t.LoadTasks()
	for i, existing := range tasks {
		if existing.ID == task.ID {
			tasks[i] = task
			t.SaveTasks(tasks)
			return
		}
	}
	t.SaveTasks(append(tasks, task))
}

// SaveTasks saves the given tasks to the repository file.
func (t TaskRepository) SaveTasks(tasks []scheduled.Task) {
	file, err := os.Create(t.path)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.path, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data := struct {
		Tasks []scheduled.Task `json:"tasks"`
	}{
		Tasks: tasks,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode tasks to %s: %v", t.path, err)
	}
}

// DeleteTask removes the task with the given id from the repository file.
func (t TaskRepository) DeleteTask(id string) {
	tasks := t.LoadTasks()
	filtered := tasks[:0]
	for _, task := range tasks {
		if task.ID != id {
			filtered = append(filtered, task)
		}
	}
	t.SaveTasks(filtered)
}

// Repository stores contexts and order in JSON files derived from the task
// file path.
type Repository struct {
	contextsPath string
	orderPath    string
}

// NewRepository creates a new Repository. taskPath must be the full path to
// the tasks .json file; context and order files are placed alongside it.
func NewRepository(path string) Repository {
	dir := filepath.Dir(path)
	stem := strings.TrimSuffix(filepath.Base(path), ".json")
	return Repository{
		contextsPath: filepath.Join(dir, stem+".contexts.json"),
		orderPath:    filepath.Join(dir, stem+".order.json"),
	}
}

// LoadContexts loads and returns all contexts from the repository file.
func (t Repository) LoadContexts() []scheduled.Context {
	file, err := os.Open(t.contextsPath)
	if err != nil {
		return []scheduled.Context{scheduled.ContextNone}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var contexts struct {
		Contexts []scheduled.Context `json:"contexts"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&contexts); err != nil {
		return []scheduled.Context{scheduled.ContextNone}
	}

	// add hard coded none context
	allContexts := []scheduled.Context{scheduled.ContextNone}
	for _, c := range contexts.Contexts {
		if c.ID != scheduled.ContextNone.ID {
			allContexts = append(allContexts, c)
		}
	}

	return allContexts
}

// LoadOrder loads task orders from the order file. If the order file does not
// exist, it migrates positions from the tasks file and saves the result.
func (t Repository) LoadOrder() []scheduled.TaskOrder {
	file, err := os.Open(t.orderPath)
	if err != nil {
		return []scheduled.TaskOrder{}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var data struct {
		Orders []scheduled.TaskOrder `json:"orders"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		log.Printf("Failed to decode %s: %v", t.orderPath, err)
		return []scheduled.TaskOrder{}
	}

	return data.Orders
}

// SaveOrder saves task orders to the order file.
func (t Repository) SaveOrder(orders []scheduled.TaskOrder) {
	file, err := os.Create(t.orderPath)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.orderPath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data := struct {
		Orders []scheduled.TaskOrder `json:"orders"`
	}{
		Orders: orders,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode orders to %s: %v", t.orderPath, err)
	}
}

// SaveContexts saves the given contexts to the repository file.
func (t Repository) SaveContexts(contexts []scheduled.Context) {
	file, err := os.Create(t.contextsPath)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.contextsPath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data := struct {
		Contexts []scheduled.Context `json:"contexts"`
	}{}
	for _, c := range contexts {
		if c.ID != 1 {
			data.Contexts = append(data.Contexts, c)
		}
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf(
			"Failed to encode contexts to %s: %v",
			t.contextsPath, err,
		)
	}
}
