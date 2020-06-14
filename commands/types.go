package commands

// Command is an generic interface for all command types
type Command interface {
	// Execute will start command logic
	Execute()
}

type baseCommand struct {
	MessageID int
	UserID    int
	Command   string
	Args      string
}

// StartCommand is an introduction command, will register this user in db
type StartCommand struct {
	Base baseCommand
}

// AddCommand will add new feed to listen
type AddCommand struct {
	Base baseCommand
}

// RemoveCommand will remove feed from listen
type RemoveCommand struct {
	Base baseCommand
}

// ListCommand will print all saved feeds
type ListCommand struct {
	Base baseCommand
}

// ReadCommand will read new posts in saved feeds
type ReadCommand struct {
	Base baseCommand
}
