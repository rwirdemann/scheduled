package file

import (
	"encoding/json"
	"log"
	"os"
	"path"

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
