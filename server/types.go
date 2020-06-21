package server

// Options holds all necessary settings for the app
type Options struct {
	Token      string
	Connection string
	Debug      bool
}

// Feed represents feed db table structure
type Feed struct {
	ID         int
	Name       string
	Normalized string
	URI        string
}
