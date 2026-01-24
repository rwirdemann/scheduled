# Scheduled

Scheduled is a TUI based rolling task manager that focus on a single work week. 

![Scheduled](scheduled.png)

# Installation

Since contexts are oaded from a JSON file, you need to copy and adapt the sample context file to $HOME/.scheduled as follows:

```bash
mkdir $HOME/.scheduled
cp contexts.sample.json $HOME/.scheduled/contexts.json
```

## MacOS and Linux

```
make install
```

Installs `scheduled` to $GOPATH/bin.

## Binary Downloads

- [MacOS Apple Silicon](bin/scheduled-macos-amd64)
- [MacOS Intel](bin/scheduled-macos-arm64)
- [Linux x86](bin/scheduled-linux-amd64)
- [Linux arm](bin/scheduled-linux-arm64)
- [Windows x86](bin/scheduled-windows-amd64.exe)

## Usage

```
./scheduled-{your-os}
```

Enter ? to toggle help.

## Libraries

Scheduled uses [Nestiles](https://github.com/rwirdemann/nestiles) for tiles management.

## Roadmap

### Feature: Task Pinning 

Normal tasks are rolling, thus a task that is scheduled for Monday will stay on Monday even if you switch weeks. Pinned tasks have a date assigned and will only appear on that specific day.

### Feature: Zoom

Zooms into a spefic day. The zoom window is a split view that consists the task list on the left and the details of the selected tasks on the right hand site. 

### Feature: Contexts

Contexts allows to switch between tasks of different contexts, like "work" or "private". The TUI should allow to switch between contexts and shows only tasks assigned to the selected context. I no context is selected all tasks are shown.

### Feature: Task Popup

Hitting <enter> on a task opens a task popup with title and decription field. The popup allows to edit the tasks title and its description. Esc quits the popup without saving, save updates the task in the underlying repository. The implementation shoud use this [Bubble](https://github.com/rmhubbert/bubbletea-overlay).

## License

* [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

