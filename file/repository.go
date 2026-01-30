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
}

// NewRepository creates a new Repository instance.
func NewRepository(filenameTasks string) Repository {
	if filenameTasks == "" {
		filenameTasks = "tasks.json"
	}
	filenameContexts := strings.TrimSuffix(filenameTasks, ".json") + ".contexts.json"
	return Repository{filenameTasks: filenameTasks, filenameContexts: filenameContexts}
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
