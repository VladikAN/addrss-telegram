package server

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/vladikan/addrss-telegram/database"
	"github.com/vladikan/addrss-telegram/templates"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	//shutdown()
	os.Exit(code)
}

func TestStart(t *testing.T) {
	exp := "start-success"
	r, err := (&Command{}).start()
	assertTemplate(t, r, exp, err)
}

func TestHelp(t *testing.T) {
	exp := "help-success"
	r, err := (&Command{}).help()
	assertTemplate(t, r, exp, err)
}

func TestAdd_NoArgs(t *testing.T) {
	exp := "add-validation"
	r, err := (&Command{}).add()
	assertTemplate(t, r, exp, err)
}

func TestAdd_ErrorOnReadActiveSubscription(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, exp },
	}

	r, err := (&Command{Args: []string{"URI"}}).add()
	assertError(t, r, err, exp)
}

func TestAdd_FeedAlreadyExists(t *testing.T) {
	exp := "add-exists"
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
	}

	r, err := (&Command{Args: []string{"URI"}}).add()
	assertTemplate(t, r, exp, err)
}

func TestAdd_ErrorOnGetFeed(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return nil, exp },
	}

	r, err := (&Command{Args: []string{"URI"}}).add()
	assertError(t, r, err, exp)
}

func TestAdd_ErrorOnSubscribe(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return &database.Feed{}, nil },
		subscribeMock:      func() error { return exp },
	}

	r, err := (&Command{Args: []string{"URI"}}).add()
	assertError(t, r, err, exp)
}

func TestAdd_SubscribeToExisting(t *testing.T) {
	exp := "add-success"
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return &database.Feed{}, nil },
		subscribeMock:      func() error { return nil },
	}

	r, err := (&Command{Args: []string{"URI"}}).add()
	assertTemplate(t, r, exp, err)
}

func TestRemove_NoArgs(t *testing.T) {
	exp := "remove-validation"
	r, err := (&Command{}).remove()
	assertTemplate(t, r, exp, err)
}

func TestRemove_ErrorOnGetNormalized(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return nil, exp },
	}

	r, err := (&Command{Args: []string{"name"}}).remove()
	assertError(t, r, err, exp)
}

func TestRemove_NoRowsToRemove(t *testing.T) {
	exp := "remove-no-rows"
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return nil, nil },
	}

	r, err := (&Command{Args: []string{"name"}}).remove()
	assertTemplate(t, r, exp, err)
}

func TestRemove_ErrorOnUnsubscribe(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
		unsubscribeMock:           func() error { return exp },
	}

	r, err := (&Command{Args: []string{"name"}}).remove()
	assertError(t, r, err, exp)
}

func TestRemove_Unsubscribed(t *testing.T) {
	exp := "remove-success"
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
		unsubscribeMock:           func() error { return nil },
	}

	r, err := (&Command{Args: []string{"name"}}).remove()
	assertTemplate(t, r, exp, err)
}

func TestList_ErrorOnRead(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserFeedsMock: func() ([]database.Feed, error) {
			return nil, exp
		},
	}

	r, err := (&Command{}).list()
	assertError(t, r, err, exp)
}

func TestList_EmptyFeeds(t *testing.T) {
	exp := "list-empty"
	db = &dbMock{
		getUserFeedsMock: func() ([]database.Feed, error) {
			return []database.Feed{}, nil
		},
	}

	r, err := (&Command{}).list()
	assertTemplate(t, r, exp, err)
}

func TestList_ListFeeds(t *testing.T) {
	exp := "list-result"
	db = &dbMock{
		getUserFeedsMock: func() ([]database.Feed, error) {
			return []database.Feed{{ID: 1}}, nil
		},
	}

	r, err := (&Command{}).list()
	assertTemplate(t, r, exp, err)
}

func assertError(t *testing.T, name string, err error, exp error) {
	if err != exp {
		t.Errorf("Expected error '%s', but was '%s'", exp, err)
	}

	if len(name) != 0 {
		t.Errorf("Expected empty response string, but was '%s'", name)
	}
}

func assertTemplate(t *testing.T, name string, exp string, err error) {
	if err != nil {
		t.Errorf("Error was not expected, but was '%s'", err)
	}

	if name != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, name)
	}
}

func setup() {
	custom := func(lang string, name string, data interface{}) (string, error) {
		return name, nil
	}

	templates.SetCustomOutput(custom)
}

type dbMock struct {
	addFeedMock               func() (*database.Feed, error)
	subscribeMock             func() error
	unsubscribeMock           func() error
	deleteUserMock            func() error
	getUserFeedsMock          func() ([]database.Feed, error)
	getUserURIFeedMock        func() (*database.Feed, error)
	getUserNormalizedFeedMock func() (*database.Feed, error)
	getFeedMock               func() (*database.Feed, error)
	getForUpdateMock          func() ([]database.Feed, error)
	getFeedUsersMock          func() ([]database.UserFeed, error)
	setFeedUpdatedMock        func() error
	setFeedLastPubMock        func() error
	setFeedBrokenMock         func() error
}

func (db *dbMock) Close() {}
func (db *dbMock) AddFeed(name string, normalized string, uri string) (*database.Feed, error) {
	return db.addFeedMock()
}
func (db *dbMock) Subscribe(userID int64, feedID int) error           { return db.subscribeMock() }
func (db *dbMock) Unsubscribe(userID int64, feedID int) error         { return db.unsubscribeMock() }
func (db *dbMock) DeleteUser(userID int64) error                      { return db.deleteUserMock() }
func (db *dbMock) GetUserFeeds(userID int64) ([]database.Feed, error) { return db.getUserFeedsMock() }
func (db *dbMock) GetUserURIFeed(userID int64, uri string) (*database.Feed, error) {
	return db.getUserURIFeedMock()
}
func (db *dbMock) GetUserNormalizedFeed(userID int64, normalized string) (*database.Feed, error) {
	return db.getUserNormalizedFeedMock()
}
func (db *dbMock) GetFeed(uri string) (*database.Feed, error)           { return db.getFeedMock() }
func (db *dbMock) GetForUpdate(count int) ([]database.Feed, error)      { return db.getForUpdateMock() }
func (db *dbMock) GetFeedUsers(feedID int) ([]database.UserFeed, error) { return db.getFeedUsersMock() }
func (db *dbMock) SetFeedUpdated(id int) error                          { return db.setFeedUpdatedMock() }
func (db *dbMock) SetFeedLastPub(id int, lastPub time.Time) error       { return db.setFeedLastPubMock() }
func (db *dbMock) SetFeedBroken(id int) error                           { return db.setFeedBrokenMock() }
