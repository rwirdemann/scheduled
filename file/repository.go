package file

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/rwirdemann/scheduled"
)

var base = "tasks.json"

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	base = home + "/.scheduled/"

	// make sure directory exists
	err = os.MkdirAll(base, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create empty context file if it does not exist
	_, err = os.Stat(path.Join(base, "contexts.json"))
	if os.IsNotExist(err) {
		destination, err := os.Create(path.Join(base, "contexts.json"))
		if err != nil {
			log.Fatal(err)
		}
		defer destination.Close()
	}
}

type Repository struct {
	filename string
}

func NewRepository(filename string) Repository {
	if filename == "" {
		filename = "tasks.json"
	}
	return Repository{filename: filename}
}

func (t Repository) LoadContexts() []scheduled.Context {
	file, err := os.Open(path.Join(base, "contexts.json"))
	if err != nil {
		log.Printf("Failed to open %s: %v", "contexts.json", err)
		return []scheduled.Context{scheduled.ContextNone}
	}
	defer file.Close()

	var contexts struct {
		Contexts []scheduled.Context `json:"contexts"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&contexts); err != nil {
		log.Printf("Failed to decode %s: %v", "contexts.json", err)
		return []scheduled.Context{scheduled.ContextNone}
	}

	// add none hard coded none context
	allContexts := []scheduled.Context{scheduled.ContextNone}
	for _, c := range contexts.Contexts {
		if c.ID != 1 {
			allContexts = append(allContexts, c)
		}
	}

	return allContexts
}

func (t Repository) Load() []scheduled.Task {
	file, err := os.Open(path.Join(base, t.filename))
	if err != nil {
		log.Printf("Failed to open %s: %v", t.filename, err)
		return []scheduled.Task{}
	}
	defer file.Close()

	var tasks struct {
		Tasks []scheduled.Task `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		log.Printf("Failed to decode %s: %v", t.filename, err)
		return []scheduled.Task{}
	}

	for i := range tasks.Tasks {
		if tasks.Tasks[i].ID == "" {
			tasks.Tasks[i].ID = uuid.NewString()
		}
	}

	return tasks.Tasks
}
func (t Repository) Save(tasks []scheduled.Task) {
	file, err := os.Create(path.Join(base, t.filename))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filename, err)
	}
	defer file.Close()

	data := struct {
		Tasks []scheduled.Task `json:"tasks"`
	}{
		Tasks: tasks,
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		log.Fatalf("Failed to encode tasks to %s: %v", t.filename, err)
	}
}
