package scheduled

// InputPort is implemented by any component that can accept a new task
// from an external input source.
type InputPort interface {
	AddTask(name string)
}

// TelegramTaskMsg is sent from the Telegram poller to the Bubble Tea
// program via p.Send(). Update() switches on this type.
type TelegramTaskMsg struct {
	Name string
}
