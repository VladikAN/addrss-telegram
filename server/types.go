package server

// Options holds all necessary settings for the app
type Options struct {
	Token        string
	DbConnection string
	Debug        bool
}

// Feed represents feed db table structure
type Feed struct {
	Name string
	URI  string
}
