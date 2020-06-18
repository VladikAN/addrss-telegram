package server

// Options holds all necessary settings for the app
type Options struct {
	Token      string
	Connection string
	Debug      bool
}

// Feed represents feed db table structure
type Feed struct {
	Name string `db:"name"`
	URI  string `db:"uri"`
}
