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

	// make sure directory exists
	err = os.MkdirAll(base, 0755)
	if err != nil {
		log.Fatal(err)
	}
}

type Repository struct {
	filenameTasks    string
	filenameContexts string
}

func NewRepository(filenameTasks string) Repository {
	if filenameTasks == "" {
		filenameTasks = "tasks.json"
	}
	filenameContexts := strings.TrimSuffix(filenameTasks, ".json") + ".contexts.json"
	return Repository{filenameTasks: filenameTasks, filenameContexts: filenameContexts}
}

func (t Repository) LoadContexts() []scheduled.Context {
	file, err := os.Open(path.Join(base, t.filenameContexts))
	if err != nil {
		return []scheduled.Context{scheduled.ContextNone}
	}
	defer file.Close()

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

func (t Repository) Load() []scheduled.Task {
	file, err := os.Open(path.Join(base, t.filenameTasks))
	if err != nil {
		return []scheduled.Task{}
	}
	defer file.Close()

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

func (t Repository) Save(tasks []scheduled.Task) {
	file, err := os.Create(path.Join(base, t.filenameTasks))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filenameTasks, err)
	}
	defer file.Close()

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

func (t Repository) SaveContexts(contexts []scheduled.Context) {
	file, err := os.Create(path.Join(base, t.filenameContexts))
	if err != nil {
		log.Fatalf("Failed to create %s: %v", t.filenameContexts, err)
	}
	defer file.Close()

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
