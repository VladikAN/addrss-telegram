package parser

import (
	"testing"
	"time"
)

func TestGetLastWithEmpty(t *testing.T) {
	var topics []Topic
	result := GetLast(topics)
	if result != nil {
		t.Errorf("Expected to be empty")
	}
}
func TestGetLastWithSingle(t *testing.T) {
	now := time.Now()
	topics := []Topic{
		{Title: "1", Date: &now},
	}
	result := GetLast(topics)
	if result == nil {
		t.Errorf("Expected to be not nil result")
	}

	if result.Title != "1" {
		t.Errorf("Expected to be title '1', but was '%s'", result.Title)
	}
}

func TestGetLastWithMany(t *testing.T) {
	now := time.Now()
	before := now.Add(-1 * time.Hour)
	after := now.Add(1 * time.Hour)

	topics := []Topic{
		{Title: "1", Date: &before},
		{Title: "2", Date: &after},
		{Title: "3", Date: &now},
	}
	result := GetLast(topics)
	if result == nil {
		t.Errorf("Expected to be not nil result")
	}

	if result.Title != "2" {
		t.Errorf("Expected to be title '2', but was '%s'", result.Title)
	}
}
