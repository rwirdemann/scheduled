# Scheduled

Scheduled is a TUI-based rolling task manager that focuses on a single work week. Tasks are added to the inbox or to the selected weekday, respectively. Tasks are moved by keyboard from day to day and stay there until they are deleted. Tasks move from week to week, scheduled for the same weekday, till they are deleted. This is especially useful for recurring tasks or tasks that haven't been finished. Contexts enable different working contexts and filter tasks based on their assigned context.

https://github.com/user-attachments/assets/fa50078e-ee0c-4e11-813f-6d196aa11b7f

# Binary Download

Download the binary for your OS from the asset section of the [latest release](https://github.com/rwirdemann/scheduled/releases/tag/0.2.1). Open a terminal and run


```
./scheduled-{your-os}
```

Enter ? to toggle help.

### Where are my tasks stored?

Tasks and contexts are stored as JSON in `$HOME/.scheduled`. The default name of the task file is `$HOME/.scheduled/tasks.json` , the name of the context file is `$HOME/.scheduled/tasks.contexts.json`. The task file name can be overriden by CLI flag `-f`. The name of the context file is derived from the tasks file. Thus, every tasks file has a dedicated set of accociated contexts. 

## Development

```
make install
```

Installs `scheduled` to $GOPATH/bin.

## Libraries

Scheduled uses [Nestiles](https://github.com/rwirdemann/nestiles) for tiles management.

## Roadmap

### Feature: Task Pinning 

Normal tasks are rolling, thus a task that is scheduled for Monday will stay on Monday even if you switch weeks. Pinned tasks have a date assigned and will only appear on that specific day.

## License

* [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

