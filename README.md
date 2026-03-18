# Scheduled

Scheduled is a TUI-based rolling task manager that focuses on a single work week. Tasks are added to the inbox or to the selected weekday, respectively. Tasks are moved by keyboard from day to day and stay there until they are deleted. Tasks move from week to week, scheduled for the same weekday, till they are deleted. This is especially useful for recurring tasks or tasks that haven't been finished. Contexts enable different working contexts and filter tasks based on their assigned context.

https://github.com/user-attachments/assets/fa50078e-ee0c-4e11-813f-6d196aa11b7f

# Binary Download

Download the binary for your OS from the asset section of the [latest release](https://github.com/rwirdemann/scheduled/releases/tag/0.2.1). Open a terminal and run


```
chmod u+x scheduled-{your-os}
./scheduled-{your-os}
```

Enter ? to toggle help.

### Where are my tasks stored?

Tasks and contexts are stored as JSON in `$HOME/.scheduled`. The default name of the task file is `$HOME/.scheduled/tasks.json` , the name of the context file is `$HOME/.scheduled/tasks.contexts.json`. The task file name can be overriden by CLI flag `-f`. The name of the context file is derived from the tasks file. Thus, every tasks file has a dedicated set of accociated contexts. 

The task and context files are saved when the application exits and every 15 seconds in the background.

## Telegram Integration

Tasks can be added to the inbox by sending a message to a Telegram bot.

### Setup

1. Create a bot via [@BotFather](https://t.me/BotFather) on Telegram and copy the API token.

2. Create a `.env` file in `$HOME/.scheduled/`:

   ```
   TELEGRAM_BOT_TOKEN=your-token-here
   ```

3. Start Scheduled as usual. The bot starts polling automatically in the background. If `TELEGRAM_BOT_TOKEN` is not set, the integration is silently disabled.

4. Send any text message to your bot — it will appear as a new task in the inbox within a few seconds.

### Scheduling tasks to a specific day

Prefix the message with a weekday name to add the task directly to that day's list instead of the inbox. Both German and English names are supported, long form and abbreviated:

| Prefix | Day |
|---|---|
| `Montag` / `Mo` / `Monday` / `Mon` | Monday |
| `Dienstag` / `Di` / `Tuesday` / `Tue` | Tuesday |
| `Mittwoch` / `Mi` / `Wednesday` / `Wed` | Wednesday |
| `Donnerstag` / `Do` / `Thursday` / `Thu` | Thursday |
| `Freitag` / `Fr` / `Friday` / `Fri` | Friday |
| `Samstag` / `Sa` / `Saturday` / `Sat` | Saturday |
| `Sonntag` / `So` / `Sunday` / `Sun` | Sunday |

**Examples:**

```
Mo Projektplanung fertigstellen
```
→ Added to Monday as "Projektplanung fertigstellen"

```
Do Review vorbereiten
```
→ Added to Thursday as "Review vorbereiten"

Messages without a weekday prefix are added to the inbox as before.

> **Note:** The `.env` file contains a secret. Never commit it to version control.

## Development

```bash
make install
```

Installs `scheduled` to $GOPATH/bin.

## Testing

```bash
go test ./...
```

## Libraries

Scheduled uses [Nestiles](https://github.com/rwirdemann/nestiles) for tiles management.

## Roadmap

### Feature: Task Pinning 

Normal tasks are rolling, thus a task that is scheduled for Monday will stay on Monday even if you switch weeks. Pinned tasks have a date assigned and will only appear on that specific day.

## License

* [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

