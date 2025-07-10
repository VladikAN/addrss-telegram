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

func TestStats_EmptyForNonAdmin(t *testing.T) {
	r, err := (&Command{}).stats()
	assertTemplate(t, r, "cmd-unknown", err)
}

func TestStats_ErrorOnQuery(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getStatsMock: func() (*database.Stats, error) { return nil, exp },
	}

	r, err := (&Command{admin: true}).stats()
	assertError(t, r, err, exp)
}

func TestStats_Success(t *testing.T) {
	db = &dbMock{
		getStatsMock: func() (*database.Stats, error) { return &database.Stats{Users: 1, Feeds: 2}, nil },
	}

	r, err := (&Command{admin: true}).stats()
	assertTemplate(t, r, "stats-success", err)
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

	r, err := (&Command{args: "URI"}).add()
	assertError(t, r, err, exp)
}

func TestAdd_FeedAlreadyExists(t *testing.T) {
	exp := "add-exists"
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
	}

	r, err := (&Command{args: "URI"}).add()
	assertTemplate(t, r, exp, err)
}

func TestAdd_ErrorOnGetFeed(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return nil, exp },
	}

	r, err := (&Command{args: "URI"}).add()
	assertError(t, r, err, exp)
}

func TestAdd_ErrorOnSubscribe(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return &database.Feed{}, nil },
		resetFeedMock:      func() error { return nil },
		subscribeMock:      func() error { return exp },
	}

	r, err := (&Command{args: "URI"}).add()
	assertError(t, r, err, exp)
}

func TestAdd_SubscribeToExisting(t *testing.T) {
	exp := "add-success"
	db = &dbMock{
		getUserURIFeedMock: func() (*database.Feed, error) { return nil, nil },
		getFeedMock:        func() (*database.Feed, error) { return &database.Feed{}, nil },
		resetFeedMock:      func() error { return nil },
		subscribeMock:      func() error { return nil },
	}

	r, err := (&Command{args: "URI"}).add()
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

	r, err := (&Command{args: "name"}).remove()
	assertError(t, r, err, exp)
}

func TestRemove_NoRowsToRemove(t *testing.T) {
	exp := "remove-no-rows"
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return nil, nil },
	}

	r, err := (&Command{args: "name"}).remove()
	assertTemplate(t, r, exp, err)
}

func TestRemove_ErrorOnUnsubscribe(t *testing.T) {
	exp := errors.New("test")
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
		unsubscribeMock:           func() error { return exp },
	}

	r, err := (&Command{args: "name"}).remove()
	assertError(t, r, err, exp)
}

func TestRemove_Unsubscribed(t *testing.T) {
	exp := "remove-success"
	db = &dbMock{
		getUserNormalizedFeedMock: func() (*database.Feed, error) { return &database.Feed{}, nil },
		unsubscribeMock:           func() error { return nil },
	}

	r, err := (&Command{args: "name"}).remove()
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

// Feedback command tests
func TestFeedback_NoArgs(t *testing.T) {
	exp := "feedback-validation"
	cmd := createTestCommand()
	replies := cmd.feedbackMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestFeedback_TooLong(t *testing.T) {
	// Create a message longer than 1000 characters
	longMessage := ""
	for i := 0; i < 1001; i++ {
		longMessage += "a"
	}
	
	exp := "feedback-too-long"
	cmd := createTestCommand()
	cmd.args = longMessage
	replies := cmd.feedbackMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestFeedback_Success(t *testing.T) {
	exp := "feedback-success"
	cmd := createTestCommand()
	cmd.args = "This is test feedback"
	cmd.userID = 12345
	cmd.adminID = 99999

	replies := cmd.feedbackMulti()
	// First reply is to admin, second is to user
	if replies[0].ChatID != 99999 {
		t.Errorf("Expected ChatID 99999, got %d", replies[0].ChatID)
	}
	// Check that admin receives the feedback-message template
	if replies[0].Text != "feedback-message" {
		t.Errorf("Expected feedback-message template, got '%s'", replies[0].Text)
	}
	assertReplyTemplate(t, replies[1], exp)
}

func TestFeedback_EmptyString(t *testing.T) {
	exp := "feedback-validation"
	cmd := createTestCommand()
	cmd.args = ""
	replies := cmd.feedbackMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestFeedback_WhitespaceOnly(t *testing.T) {
	exp := "feedback-validation"
	cmd := createTestCommand()
	cmd.args = "   "
	replies := cmd.feedbackMulti()
	assertReplyTemplate(t, replies[0], exp)
}

// Notify command tests
func TestNotify_NonAdmin(t *testing.T) {
	exp := "cmd-unknown"
	cmd := createTestCommand()
	cmd.admin = false
	cmd.args = "Test notification"
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestNotify_NoArgs(t *testing.T) {
	exp := "notify-validation"
	cmd := createTestCommand()
	cmd.admin = true
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestNotify_TooLong(t *testing.T) {
	// Create a message longer than 2000 characters
	longMessage := ""
	for i := 0; i < 2001; i++ {
		longMessage += "a"
	}
	
	exp := "notify-too-long"
	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = longMessage
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestNotify_ErrorOnGetUsers(t *testing.T) {
	exp := "notify-error"
	db = &dbMock{
		getAllUsersMock: func() ([]int64, error) {
			return nil, errors.New("database error")
		},
	}

	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = "Test notification"
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[len(replies)-1], exp)
}

func TestNotify_Success(t *testing.T) {
	exp := "notify-success"
	userIDs := []int64{123, 456, 789}
	db = &dbMock{
		getAllUsersMock: func() ([]int64, error) {
			return userIDs, nil
		},
	}

	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = "Test notification"

	replies := cmd.notifyMulti()
	// The last reply is the summary to the admin
	assertReplyTemplate(t, replies[len(replies)-1], exp)
	// The rest should be notifications to users
	for i, userID := range userIDs {
		reply := replies[i]
		if reply.ChatID != userID {
			t.Errorf("Expected ChatID %d, got %d", userID, reply.ChatID)
		}
		// Check that users receive the notify-message template
		if reply.Text != "notify-message" {
			t.Errorf("Expected notify-message template, got '%s'", reply.Text)
		}
	}
}

func TestNotify_EmptyUsers(t *testing.T) {
	exp := "notify-success"
	db = &dbMock{
		getAllUsersMock: func() ([]int64, error) {
			return []int64{}, nil
		},
	}

	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = "Test notification"
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestNotify_EmptyString(t *testing.T) {
	exp := "notify-validation"
	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = ""
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func TestNotify_WhitespaceOnly(t *testing.T) {
	exp := "notify-validation"
	cmd := createTestCommand()
	cmd.admin = true
	cmd.args = "   "
	replies := cmd.notifyMulti()
	assertReplyTemplate(t, replies[0], exp)
}

func assertError(t *testing.T, resp string, err error, exp error) {
	if err != exp {
		t.Errorf("Expected error '%s', but was '%s'", exp, err)
	}

	if len(resp) != 0 {
		t.Errorf("Expected empty response string, but was '%s'", resp)
	}
}

func assertTemplate(t *testing.T, resp string, exp string, err error) {
	if err != nil {
		t.Errorf("Error was not expected, but was '%s'", err)
	}

	if resp != exp {
		t.Errorf("Expected '%s', but was '%s'", exp, resp)
	}
}

func assertReplyTemplate(t *testing.T, reply Reply, exp string) {
	if reply.Text != exp {
		t.Errorf("Expected reply template '%s', got '%s'", exp, reply.Text)
	}
}

func setup() {
	custom := func(lang string, name string, data interface{}) (string, error) {
		return name, nil
	}

	templates.SetCustomOutput(custom)
}

func createTestCommand() *Command {
	return &Command{
		replyQueue: make(chan Reply, 10),
	}
}

type dbMock struct {
	getStatsMock              func() (*database.Stats, error)
	addFeedMock               func() (*database.Feed, error)
	subscribeMock             func() error
	unsubscribeMock           func() error
	deleteUserMock            func() error
	getUserFeedsMock          func() ([]database.Feed, error)
	getUserURIFeedMock        func() (*database.Feed, error)
	getUserNormalizedFeedMock func() (*database.Feed, error)
	getFeedMock               func() (*database.Feed, error)
	getFeedsMock              func() ([]database.Feed, error)
	resetFeedMock             func() error
	getFeedUsersMock          func() ([]database.UserFeed, error)
	getAllUsersMock           func() ([]int64, error)
	setFeedUpdatedMock        func() error
	setFeedLastPubMock        func() error
	setFeedBrokenMock         func() error
}

func (db *dbMock) Close()                             {}
func (db *dbMock) GetStats() (*database.Stats, error) { return db.getStatsMock() }
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
func (db *dbMock) GetFeeds(count int) ([]database.Feed, error)          { return db.getFeedsMock() }
func (db *dbMock) GetFeedUsers(feedID int) ([]database.UserFeed, error) { return db.getFeedUsersMock() }
func (db *dbMock) GetAllUsers() ([]int64, error)                        { return db.getAllUsersMock() }
func (db *dbMock) ResetFeed(feedID int) error                           { return db.resetFeedMock() }
func (db *dbMock) SetFeedUpdated(id int) error                          { return db.setFeedUpdatedMock() }
func (db *dbMock) SetFeedLastPub(id int, lastPub time.Time, lastPubURI string) error { return db.setFeedLastPubMock() }
func (db *dbMock) SetFeedBroken(id int) error                           { return db.setFeedBrokenMock() }
