package scheduled

import (
	"fmt"
	"slices"
	"sort"

	"github.com/google/uuid"
)

type Plan struct {
	tasks  []Task
	orders map[string]int
}

// NewPlan creates a Plan from tasks and their display orders. It normalizes
// all positions on creation.
func NewPlan(tasks []Task, orders []TaskOrder) *Plan {
	p := &Plan{
		tasks:  slices.Clone(tasks),
		orders: make(map[string]int),
	}
	for _, o := range orders {
		p.orders[o.TaskID] = o.Pos
	}
	p.normalizeAllPositions()
	p.sort(p.tasks)
	return p
}

// AllTasks returns all tasks sorted by day and position.
func (p *Plan) AllTasks() []Task {
	return slices.Clone(p.tasks)
}

// AllOrders returns the current task orders as a slice for persistence.
func (p *Plan) AllOrders() []TaskOrder {
	orders := make([]TaskOrder, 0, len(p.orders))
	for id, pos := range p.orders {
		orders = append(orders, TaskOrder{TaskID: id, Pos: pos})
	}
	return orders
}

// TasksForDay returns all tasks for the given day, sorted by position.
func (p *Plan) TasksForDay(day int) []Task {
	var result []Task
	for _, task := range p.tasks {
		if task.Day == day {
			result = append(result, task)
		}
	}
	return result
}

// TasksForDayAndContext returns tasks for the given day filtered by context,
// sorted by position.
func (p *Plan) TasksForDayAndContext(day int, contextID int) []Task {
	if contextID == ContextNone.ID {
		return p.TasksForDay(day)
	}

	var result []Task
	for _, task := range p.tasks {
		if task.Day == day && task.Context == contextID {
			result = append(result, task)
		}
	}
	return result
}

// CreateTask creates a new task and appends it to the given day.
func (p *Plan) CreateTask(name string, contextID int, description string, day int) Task {
	task := Task{
		ID:      uuid.NewString(),
		Name:    name,
		Desc:    description,
		Day:     day,
		Done:    false,
		Context: contextID,
	}
	p.orders[task.ID] = len(p.TasksForDay(day))
	p.tasks = append(p.tasks, task)
	p.normalizeDayPositions(day)
	p.sort(p.tasks)
	return task
}

// UpdateTask updates the name, context, and description of the task with the
// given id.
func (p *Plan) UpdateTask(id, name string, contextID int, description string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	p.tasks[index].Name = name
	p.tasks[index].Context = contextID
	p.tasks[index].Desc = description
	return nil
}

// ToggleDone toggles the done state of the task with the given id.
func (p *Plan) ToggleDone(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	p.tasks[index].Done = !p.tasks[index].Done
	return nil
}

// DeleteDoneTask removes a completed task from the plan.
func (p *Plan) DeleteDoneTask(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if !p.tasks[index].Done {
		return nil
	}

	day := p.tasks[index].Day
	delete(p.orders, id)
	p.tasks = append(p.tasks[:index], p.tasks[index+1:]...)
	p.normalizeDayPositions(day)
	p.sort(p.tasks)
	return nil
}

// MoveTaskToDay moves the task with the given id to toDay.
func (p *Plan) MoveTaskToDay(id string, toDay int) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	fromDay := p.tasks[index].Day
	if fromDay == toDay {
		return nil
	}

	p.tasks[index].Day = toDay
	p.orders[id] = len(p.TasksForDay(toDay))

	p.normalizeDayPositions(fromDay)
	p.normalizeDayPositions(toDay)
	p.sort(p.tasks)
	return nil
}

// MoveTaskUp moves the task with the given id one position up within its day.
func (p *Plan) MoveTaskUp(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	day := p.tasks[index].Day
	dayIndices := p.indicesForDay(day)
	posInDay := indexOfIndex(dayIndices, index)
	if posInDay <= 0 {
		return nil
	}

	prevIndex := dayIndices[posInDay-1]
	prevID := p.tasks[prevIndex].ID
	p.orders[id], p.orders[prevID] = p.orders[prevID], p.orders[id]
	p.normalizeDayPositions(day)
	p.sort(p.tasks)
	return nil
}

// MoveTaskDown moves the task with the given id one position down within its
// day.
func (p *Plan) MoveTaskDown(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	day := p.tasks[index].Day
	dayIndices := p.indicesForDay(day)
	posInDay := indexOfIndex(dayIndices, index)
	if posInDay < 0 || posInDay >= len(dayIndices)-1 {
		return nil
	}

	nextIndex := dayIndices[posInDay+1]
	nextID := p.tasks[nextIndex].ID
	p.orders[id], p.orders[nextID] = p.orders[nextID], p.orders[id]
	p.normalizeDayPositions(day)
	p.sort(p.tasks)
	return nil
}

// IsContextUsed reports whether any task uses the given context.
func (p *Plan) IsContextUsed(contextID int) bool {
	for _, task := range p.tasks {
		if task.Context == contextID {
			return true
		}
	}
	return false
}

func (p *Plan) indexByID(id string) int {
	for i, task := range p.tasks {
		if task.ID == id {
			return i
		}
	}
	return -1
}

func indexOfIndex(indices []int, wanted int) int {
	for i, index := range indices {
		if index == wanted {
			return i
		}
	}
	return -1
}

// sort sorts tasks by day, then by position within the day, then
// by ID as a tiebreaker.
func (p *Plan) sort(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Day != tasks[j].Day {
			return tasks[i].Day < tasks[j].Day
		}
		if p.orders[tasks[i].ID] != p.orders[tasks[j].ID] {
			return p.orders[tasks[i].ID] < p.orders[tasks[j].ID]
		}
		return tasks[i].ID < tasks[j].ID
	})
}

// normalizeAllPositions normalizes positions for every day that has tasks.
// The map collects unique day values as a set; struct{} costs zero bytes.
func (p *Plan) normalizeAllPositions() {
	days := map[int]struct{}{}
	for _, task := range p.tasks {
		days[task.Day] = struct{}{}
	}
	for day := range days {
		p.normalizeDayPositions(day)
	}
}

// normalizeDayPositions reassigns positions 0, 1, 2, … to all tasks of day
// in their current order, closing any gaps left by moves or deletions.
func (p *Plan) normalizeDayPositions(day int) {
	dayIndices := p.indicesForDay(day)
	for pos, index := range dayIndices {
		p.orders[p.tasks[index].ID] = pos
	}
}

// indicesForDay returns the indices of all tasks belonging to day, sorted by
// position and ID as a tiebreaker.
func (p *Plan) indicesForDay(day int) []int {
	var indices []int
	for i := range p.tasks {
		if p.tasks[i].Day == day {
			indices = append(indices, i)
		}
	}

	sort.Slice(indices, func(i, j int) bool {
		leftID := p.tasks[indices[i]].ID
		rightID := p.tasks[indices[j]].ID
		leftPos := p.orders[leftID]
		rightPos := p.orders[rightID]
		if leftPos == rightPos {
			return leftID < rightID
		}
		return leftPos < rightPos
	})

	return indices
}
