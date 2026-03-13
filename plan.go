package scheduled

import (
	"fmt"
	"slices"
	"sort"

	"github.com/google/uuid"
)

type Plan struct {
	tasks []Task
}

func NewPlan(tasks []Task) *Plan {
	cloned := slices.Clone(tasks)
	sortTasks(cloned)
	normalizeAllPositions(cloned)
	return &Plan{tasks: cloned}
}

func (p *Plan) AllTasks() []Task {
	cloned := slices.Clone(p.tasks)
	sortTasks(cloned)
	return cloned
}

func (p *Plan) TasksForDay(day int) []Task {
	var result []Task
	for _, task := range p.tasks {
		if task.Day == day {
			result = append(result, task)
		}
	}
	sortTasks(result)
	return result
}

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
	sortTasks(result)
	return result
}

func (p *Plan) CreateTask(name string, contextID int, description string, day int) Task {
	task := Task{
		ID:      uuid.NewString(),
		Name:    name,
		Desc:    description,
		Day:     day,
		Done:    false,
		Context: contextID,
		Pos:     len(p.TasksForDay(day)),
	}
	p.tasks = append(p.tasks, task)
	normalizeDayPositions(p.tasks, day)
	return task
}

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

func (p *Plan) ToggleDone(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	p.tasks[index].Done = !p.tasks[index].Done
	return nil
}

func (p *Plan) DeleteDoneTask(id string) error {
	index := p.indexByID(id)
	if index < 0 {
		return fmt.Errorf("task %q not found", id)
	}

	if !p.tasks[index].Done {
		return nil
	}

	day := p.tasks[index].Day
	p.tasks = append(p.tasks[:index], p.tasks[index+1:]...)
	normalizeDayPositions(p.tasks, day)
	return nil
}

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
	p.tasks[index].Pos = len(p.TasksForDay(toDay))

	normalizeDayPositions(p.tasks, fromDay)
	normalizeDayPositions(p.tasks, toDay)
	return nil
}

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
	p.tasks[index].Pos, p.tasks[prevIndex].Pos = p.tasks[prevIndex].Pos, p.tasks[index].Pos
	normalizeDayPositions(p.tasks, day)
	return nil
}

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
	p.tasks[index].Pos, p.tasks[nextIndex].Pos = p.tasks[nextIndex].Pos, p.tasks[index].Pos
	normalizeDayPositions(p.tasks, day)
	return nil
}

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

func (p *Plan) indicesForDay(day int) []int {
	var indices []int
	for i := range p.tasks {
		if p.tasks[i].Day == day {
			indices = append(indices, i)
		}
	}

	sort.Slice(indices, func(i, j int) bool {
		left := p.tasks[indices[i]]
		right := p.tasks[indices[j]]
		if left.Pos == right.Pos {
			return left.ID < right.ID
		}
		return left.Pos < right.Pos
	})

	return indices
}

func indexOfIndex(indices []int, wanted int) int {
	for i, index := range indices {
		if index == wanted {
			return i
		}
	}
	return -1
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Day != tasks[j].Day {
			return tasks[i].Day < tasks[j].Day
		}
		if tasks[i].Pos != tasks[j].Pos {
			return tasks[i].Pos < tasks[j].Pos
		}
		return tasks[i].ID < tasks[j].ID
	})
}

func normalizeAllPositions(tasks []Task) {
	days := map[int]struct{}{}
	for _, task := range tasks {
		days[task.Day] = struct{}{}
	}

	for day := range days {
		normalizeDayPositions(tasks, day)
	}
}

func normalizeDayPositions(tasks []Task, day int) {
	var dayIndices []int
	for i := range tasks {
		if tasks[i].Day == day {
			dayIndices = append(dayIndices, i)
		}
	}

	sort.Slice(dayIndices, func(i, j int) bool {
		left := tasks[dayIndices[i]]
		right := tasks[dayIndices[j]]
		if left.Pos == right.Pos {
			return left.ID < right.ID
		}
		return left.Pos < right.Pos
	})

	for pos, index := range dayIndices {
		tasks[index].Pos = pos
	}
}
