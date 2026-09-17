package session

import (
	"testing"
	"time"
)

func TestSessionManager_Lifecycle(t *testing.T) {
	sm := NewSessionManager()

	// Test Empty Completion
	turn := sm.CompleteTurn()
	if turn != nil {
		t.Errorf("Expected nil turn when nothing started, got %+v", turn)
	}

	// Test Start and Append
	sm.StartTurn("Hello")
	sm.SetCandidateResponse("Hi there!")
	sm.SetAISuggestion("Greetings")

	turn = sm.CompleteTurn()
	if turn == nil {
		t.Fatalf("Expected turn, got nil")
	}
	if turn.TurnIndex != 1 {
		t.Errorf("Expected TurnIndex 1, got %d", turn.TurnIndex)
	}
	if turn.InterviewerQuestion != "Hello" {
		t.Errorf("Expected InterviewerQuestion 'Hello', got '%s'", turn.InterviewerQuestion)
	}
	if turn.CandidateResponse != "Hi there!" {
		t.Errorf("Expected CandidateResponse 'Hi there!', got '%s'", turn.CandidateResponse)
	}
	if turn.AISuggestion != "Greetings" {
		t.Errorf("Expected AISuggestion 'Greetings', got '%s'", turn.AISuggestion)
	}
	if turn.LatencyMs < 0 {
		t.Errorf("Expected non-negative LatencyMs, got %d", turn.LatencyMs)
	}
}

func TestSessionManager_GetRecentTurns(t *testing.T) {
	sm := NewSessionManager()

	turns := sm.GetRecentTurns(2)
	if len(turns) != 0 {
		t.Errorf("Expected 0 recent turns, got %d", len(turns))
	}

	sm.StartTurn("Q1")
	sm.SetCandidateResponse("R1")
	sm.CompleteTurn()

	sm.StartTurn("Q2")
	sm.SetCandidateResponse("R2")
	sm.CompleteTurn()

	sm.StartTurn("Q3")
	sm.SetCandidateResponse("R3")
	sm.CompleteTurn()

	recent := sm.GetRecentTurns(2)
	if len(recent) != 2 {
		t.Fatalf("Expected 2 recent turns, got %d", len(recent))
	}
	if recent[0].TurnIndex != 2 || recent[1].TurnIndex != 3 {
		t.Errorf("Expected TurnIndex 2 and 3, got %d and %d", recent[0].TurnIndex, recent[1].TurnIndex)
	}

	recent = sm.GetRecentTurns(5)
	if len(recent) != 3 {
		t.Fatalf("Expected 3 recent turns when asking for more than available, got %d", len(recent))
	}
}

func TestSessionManager_ExportAndClear(t *testing.T) {
	sm := NewSessionManager()

	sm.StartTurn("Q1")
	sm.SetCandidateResponse("R1")
	sm.CompleteTurn()

	export := sm.Export()

	if _, ok := export["session_started_at"].(time.Time); !ok {
		t.Errorf("Expected session_started_at to be time.Time")
	}
	if _, ok := export["session_duration_s"].(int); !ok {
		t.Errorf("Expected session_duration_s to be int")
	}
	if export["turn_count"] != 1 {
		t.Errorf("Expected turn_count 1, got %v", export["turn_count"])
	}
	if turns, ok := export["turns"].([]Turn); !ok || len(turns) != 1 {
		t.Errorf("Expected turns slice with length 1, got %v", export["turns"])
	}

	sm.Clear()

	exportAfter := sm.Export()
	if exportAfter["turn_count"] != 0 {
		t.Errorf("Expected turn_count 0 after Clear, got %v", exportAfter["turn_count"])
	}

	recent := sm.GetRecentTurns(2)
	if len(recent) != 0 {
		t.Errorf("Expected 0 recent turns after Clear, got %d", len(recent))
	}
}

func TestSessionManager_Concurrency(t *testing.T) {
	sm := NewSessionManager()

	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			sm.StartTurn("ConQ")
			sm.SetCandidateResponse("ConR")
			sm.CompleteTurn()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			sm.GetRecentTurns(2)
			sm.Export()
		}
		done <- true
	}()

	<-done
	<-done

	if len(sm.turns) != 100 {
		t.Errorf("Expected 100 turns, got %d", len(sm.turns))
	}
}
