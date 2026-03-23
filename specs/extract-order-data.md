# Extract order data from tasks domain model

## Acceptance Criteria

### Order within a list

- A newly created task is appended at the end of its day list.
- MoveTaskUp moves a task one position up within its day; the task
  above moves one position down.
- MoveTaskDown moves a task one position down within its day; the task
  below moves one position up.
- MoveTaskUp on the first task of a day has no effect.
- MoveTaskDown on the last task of a day has no effect.
- Order is preserved across restarts (persisted via AllOrders /
  SaveOrder).

### Moving tasks between lists

- MoveTaskToDay appends the task to the end of the target day list.
- The order of remaining tasks in the source list is unchanged.
- MoveTaskToDay to the same day has no effect.
- After moving, the task exists only in the target list.

### Context filtering

- When a context is active, TasksForDayAndContext returns only tasks
  belonging to that context; tasks of other contexts are hidden.
- MoveTaskUp / MoveTaskDown operates on the full day list, not the
  filtered view. A task may jump over hidden tasks as a result.
- AllOrders includes all tasks of a day regardless of the active
  context.
- After deactivating the context filter, the global order of all tasks
  in a day is correct.
