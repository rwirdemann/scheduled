package file

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/rwirdemann/scheduled"
)

var base string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	base = home + "/.scheduled/"

	// make sure the directory exists
	err = os.MkdirAll(base, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

// Repository stores tasks and contexts in JSON files.
type Repository struct {
	filenameTasks    string
	filenameContexts string
	filenameOrder    string
}

// NewRepository creates a new Repository instance.
func NewRepository(filenameTasks string) Repository {
	if filenameTasks == "" {
		filenameTasks = "tasks.json"
	}
	filenameContexts := strings.TrimSuffix(filenameTasks, ".json") +
		".contexts.json"
	filenameOrder := strings.TrimSuffix(filenameTasks, ".json") +
		".order.json"
	return Repository{
		filenameTasks:    filenameTasks,
		filenameContexts: filenameContexts,
		filenameOrder:    filenameOrder,
	}
}

// LoadContexts loads and returns all contexts from the repository file.
func (t Repository) LoadContexts() []scheduled.Context {
	file, err := os.Open(path.Join(base, t.filenameContexts))
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

// LoadTasks loads and returns all tasks from the repository file.
func (t Repository) LoadTasks() []scheduled.Task {
	file, err := os.Open(path.Join(base, t.filenameTasks))
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
		log.Printf("Failed to decode %s: %v", t.filenameTasks, err)
		return []scheduled.Task{}
	}

	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == "" {
			tasks.Tasks[i].ID = uuid.NewString()
		}
	}

	return tasks.Tasks
}

// LoadOrder loads task orders from the order file. If the order file does not
// exist, it migrates positions from the tasks file and saves the result.
func (t Repository) LoadOrder() []scheduled.TaskOrder {
	orderPath := path.Join(base, t.filenameOrder)
	if _, err := os.Stat(orderPath); os.IsNotExist(err) {
		return t.migrateOrderFromTasks()
	}

	file, err := os.Open(orderPath)
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
		log.Printf("Failed to decode %s: %v", t.filenameOrder, err)
		return []scheduled.TaskOrder{}
	}

	return data.Orders
}

// migrateOrderFromTasks reads the legacy pos field from tasks.json, builds a
// TaskOrder slice, and saves it to the order file.
func (t Repository) migrateOrderFromTasks() []scheduled.TaskOrder {
	type taskWithPos struct {
		ID  string `json:"id"`
		Pos int    `json:"pos"`
	}

	file, err := os.Open(path.Join(base, t.filenameTasks))
	if err != nil {
		return []scheduled.TaskOrder{}
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var data struct {
		Tasks []taskWithPos `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return []scheduled.TaskOrder{}
	}

	orders := make([]scheduled.TaskOrder, 0, len(data.Tasks))
	for _, tw := range data.Tasks {
		if tw.ID != "" {
			orders = append(
				orders,
				scheduled.TaskOrder{TaskID: tw.ID, Pos: tw.Pos},
			)
		}
	}

	t.SaveOrder(orders)
	return orders
}

// SaveOrder saves task orders to the order file.
func (t Repository) SaveOrder(orders []scheduled.TaskOrder) {
	file, err := os.Create(path.Join(base, t.filenameOrder))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filenameOrder, err)
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
		log.Fatalf("Failed to encode orders to %s: %v", t.filenameOrder, err)
	}
}

// SaveTask upserts a single task in the repository file.
func (t Repository) SaveTask(task scheduled.Task) {
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

// DeleteTask removes the task with the given id from the repository file.
func (t Repository) DeleteTask(id string) {
	tasks := t.LoadTasks()
	filtered := tasks[:0]
	for _, task := range tasks {
		if task.ID != id {
			filtered = append(filtered, task)
		}
	}
	t.SaveTasks(filtered)
}

// SaveTasks saves the given tasks to the repository file.
func (t Repository) SaveTasks(tasks []scheduled.Task) {
	file, err := os.Create(path.Join(base, t.filenameTasks))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filenameTasks, err)
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
		log.Fatalf("Failed to encode tasks to %s: %v", t.filenameTasks, err)
	}
}

// SaveContexts saves the given contexts to the repository file.
func (t Repository) SaveContexts(contexts []scheduled.Context) {
	file, err := os.Create(path.Join(base, t.filenameContexts))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filenameContexts, err)
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
		log.Fatalf("Failed to encode contexts to %s: %v", t.filenameContexts, err)
	}
}
