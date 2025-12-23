package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestGetStoerungText_KnownAndUnknown(t *testing.T) {
	known := getStoerungText(1)
	if known != "Sicherung F25 defekt" {
		t.Fatalf("expected known text for 1, got %q", known)
	}

	unknown := getStoerungText(999)
	if unknown != "Unbekannte Störung" {
		t.Fatalf("expected unknown text, got %q", unknown)
	}
}

func TestHandleZRecord_SetAndQuit(t *testing.T) {
	// ensure fresh stoerungRecord for test
	stoerungRecord = newEmptyStoerungRecord(nodeStoerung)

	// Simulate a Set event
	fieldsSet := []string{"z", "18:39:41", "Stoerung", "Set", "7"}
	lineSet := strings.Join([]string{"z", "18:39:41", "Stoerung", "Set", "7", "Stop:1"}, " ")
	handleZRecord(fieldsSet, lineSet)

	if stoerungRecord.StoerungNr.Value != 7 {
		t.Fatalf("after Set expected StoerungNr 7, got %v", stoerungRecord.StoerungNr.Value)
	}
	if !stoerungRecord.StoerungActive.Value {
		t.Fatalf("after Set expected StoerungActive true, got false")
	}
	if stoerungRecord.StoerungText.Value != getStoerungText(7) {
		t.Fatalf("after Set expected StoerungText %q, got %q", getStoerungText(7), stoerungRecord.StoerungText.Value)
	}

	// Simulate a Quit event with padded number "0007"
	fieldsQuit := []string{"z", "18:40:16", "Stoerung", "Quit", "0007"}
	lineQuit := "z 18:40:16 Stoerung Quit 0007"
	handleZRecord(fieldsQuit, lineQuit)

	if stoerungRecord.StoerungNr.Value != 7 {
		t.Fatalf("after Quit expected StoerungNr 7, got %v", stoerungRecord.StoerungNr.Value)
	}
	if stoerungRecord.StoerungActive.Value {
		t.Fatalf("after Quit expected StoerungActive false, got true")
	}
	// LastActive should be set to the time field from input (fields[1])
	if stoerungRecord.LastActive.Value != "18:40:16" {
		t.Fatalf("after Quit expected LastActive %q, got %q", "18:40:16", stoerungRecord.LastActive.Value)
	}
}

// Optional helper to ensure strconv.Atoi behavior for padded numbers (ensures test expectations)
func TestAtoiPadded(t *testing.T) {
	v, err := strconv.Atoi("0007")
	if err != nil {
		t.Fatalf("Atoi failed: %v", err)
	}
	if v != 7 {
		t.Fatalf("Atoi expected 7, got %d", v)
	}
}
