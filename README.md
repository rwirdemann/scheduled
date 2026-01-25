# Scheduled

Scheduled is a TUI based rolling task manager that focus on a single work week. 

![Scheduled](scheduled.png)

## Binary Download

Download the binary for your OS from the asset section of the [latest release](https://github.com/rwirdemann/scheduled/releases/tag/0.2.1). Open a terminal and run


```
./scheduled-{your-os}
```

Enter ? to toggle help.

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

