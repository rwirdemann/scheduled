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
}

type Repository struct {
}

func (t Repository) Load() []scheduled.Task {
	file, err := os.Open(path.Join(base, "tasks.json"))
	if err != nil {
		log.Printf("Failed to open tasks.json: %v", err)
		return []scheduled.Task{}
	}
	defer file.Close()

	var tasks struct {
		Tasks []scheduled.Task `json:"tasks"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tasks); err != nil {
		log.Printf("Failed to decode tasks.json: %v", err)
		return []scheduled.Task{}
	}

	return tasks.Tasks
}
