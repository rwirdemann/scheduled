package clipboard

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rwirdemann/scheduled"
)

// FormatTasks formats a list of tasks into a string suitable for copying to the
// clipboard.
func FormatTasks(contexts []scheduled.Context, tasks []scheduled.Task) string {
	// Group tasks by context
	tasksByContext := make(map[int][]scheduled.Task)
	for _, task := range tasks {
		tasksByContext[task.Context] = append(tasksByContext[task.Context], task)
	}

	// Get context IDs and sort them
	contextIDs := make([]int, 0, len(tasksByContext))
	for contextID := range tasksByContext {
		contextIDs = append(contextIDs, contextID)
	}
	sort.Ints(contextIDs)

	// Build the output string
	var result []string
	for _, contextID := range contextIDs {
		// Find context name
		contextName := "Unknown"
		for _, ctx := range contexts {
			if ctx.ID == contextID {
				contextName = ctx.Name
				break
			}
		}

		// Get task names
		taskNames := make([]string, 0, len(tasksByContext[contextID]))
		for _, task := range tasksByContext[contextID] {
			taskNames = append(taskNames, task.Name)
		}

		// Format: "ContextName: Task1, Task2"
		result = append(result, fmt.Sprintf("%s: %s", contextName, strings.Join(taskNames, ", ")))
	}

	// Join with semicolons
	return strings.Join(result, "; ")
}
