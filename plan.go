package scheduled

import (
	"fmt"
	"slices"
	"sort"

	"github.com/google/uuid"
)

// Plan holds all tasks for a week, grouped by day. Within each day
// the slice order defines the display position.
type Plan struct {
	tasks map[int][]Task
}

// NewPlan creates a Plan from tasks and their display orders.
func NewPlan(tasks []Task, orders []TaskOrder) *Plan {
	p := &Plan{tasks: make(map[int][]Task)}

	posMap := make(map[string]int, len(orders))
	for _, o := range orders {
		posMap[o.TaskID] = o.Pos
	}

	for _, t := range tasks {
		p.tasks[t.Day] = append(p.tasks[t.Day], t)
	}

	for day := range p.tasks {
		sort.Slice(p.tasks[day], func(i, j int) bool {
			pi := posMap[p.tasks[day][i].ID]
			pj := posMap[p.tasks[day][j].ID]
			if pi != pj {
				return pi < pj
			}
			return p.tasks[day][i].ID < p.tasks[day][j].ID
		})
	}

	return p
}

// AllTasks returns all tasks sorted by day and position.
func (p *Plan) AllTasks() []Task {
	days := make([]int, 0, len(p.tasks))
	for day := range p.tasks {
		days = append(days, day)
	}
	sort.Ints(days)

	var result []Task
	for _, day := range days {
		result = append(result, p.tasks[day]...)
	}
	return result
}

// AllOrders returns the current task orders as a slice for persistence.
func (p *Plan) AllOrders() []TaskOrder {
	var orders []TaskOrder
	for _, tasks := range p.tasks {
		for pos, task := range tasks {
			orders = append(orders,
				TaskOrder{TaskID: task.ID, Pos: pos})
		}
	}
	return orders
}

// TasksForDay returns all tasks for the given day, sorted by position.
func (p *Plan) TasksForDay(day int) []Task {
	return slices.Clone(p.tasks[day])
}

// TasksForDayAndContext returns tasks for the given day filtered by
// context, sorted by position.
func (p *Plan) TasksForDayAndContext(day int, contextID int) []Task {
	if contextID == ContextNone.ID {
		return p.TasksForDay(day)
	}

	var result []Task
	for _, task := range p.tasks[day] {
		if task.Context == contextID {
			result = append(result, task)
		}
	}
	return result
}

// CreateTask creates a new task and appends it to the given day.
func (p *Plan) CreateTask(
	name string, contextID int, description string, day int,
) Task {
	task := Task{
		ID:      uuid.NewString(),
		Name:    name,
		Desc:    description,
		Day:     day,
		Done:    false,
		Context: contextID,
	}
	p.tasks[day] = append(p.tasks[day], task)
	return task
}

// UpdateTask updates the name, context, and description of the task
// with the given id.
func (p *Plan) UpdateTask(
	id, name string, contextID int, description string,
) error {
	day, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	p.tasks[day][index].Name = name
	p.tasks[day][index].Context = contextID
	p.tasks[day][index].Desc = description
	return nil
}

// ToggleDone toggles the done state of the task with the given id.
func (p *Plan) ToggleDone(id string) error {
	day, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	p.tasks[day][index].Done = !p.tasks[day][index].Done
	return nil
}

// DeleteDoneTask removes a completed task from the plan.
func (p *Plan) DeleteDoneTask(id string) error {
	day, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if !p.tasks[day][index].Done {
		return nil
	}

	p.tasks[day] = append(
		p.tasks[day][:index], p.tasks[day][index+1:]...)
	return nil
}

// MoveTaskToDay moves the task with the given id to toDay.
func (p *Plan) MoveTaskToDay(id string, toDay int) error {
	fromDay, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if fromDay == toDay {
		return nil
	}

	task := p.tasks[fromDay][index]
	task.Day = toDay
	p.tasks[fromDay] = append(
		p.tasks[fromDay][:index], p.tasks[fromDay][index+1:]...)
	p.tasks[toDay] = append(p.tasks[toDay], task)
	return nil
}

// MoveTaskUp moves the task with the given id one position up within
// its day.
func (p *Plan) MoveTaskUp(id string) error {
	day, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if index == 0 {
		return nil
	}

	p.tasks[day][index], p.tasks[day][index-1] =
		p.tasks[day][index-1], p.tasks[day][index]
	return nil
}

// MoveTaskDown moves the task with the given id one position down
// within its day.
func (p *Plan) MoveTaskDown(id string) error {
	day, index := p.findTask(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if index >= len(p.tasks[day])-1 {
		return nil
	}

	p.tasks[day][index], p.tasks[day][index+1] =
		p.tasks[day][index+1], p.tasks[day][index]
	return nil
}

// IsContextUsed reports whether any task uses the given context.
func (p *Plan) IsContextUsed(contextID int) bool {
	for _, tasks := range p.tasks {
		for _, task := range tasks {
			if task.Context == contextID {
				return true
			}
		}
	}
	return false
}

// findTask returns the day and slice index of the task with the given
// id, or (-1, -1) if not found.
func (p *Plan) findTask(id string) (day, index int) {
	for day, tasks := range p.tasks {
		for i, task := range tasks {
			if task.ID == id {
				return day, i
			}
		}
	}
	return -1, -1
}
