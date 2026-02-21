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

func TestHandleZRecord_Duration(t *testing.T) {
	// 026/02/14 13:21:44 Handling Z record: fields:[z|14:10:40|Kessel|Zündung] <-- Hier beginnt die Zündung
	// 2026/02/14 13:31:25 Handling Z record: fields:[z|14:20:20|Kessel|Leistungsbrand] <-- Hier beginnt der Leistungsbrand
	// 2026/02/14 17:11:37 Handling Z record: fields:[z|18:00:32|Kessel|Aus] <-- Leistungsbrand endet

	kesselRecord = newEmptyKesselRecord(nodeKessel)

	// Start Zündung
	handleZRecord([]string{"z", "14:10:40", "Kessel", "Zündung"}, "z 14:10:40 Kessel Zündung")
	if kesselRecord.AnzahlZuendungen.Value != 1 {
		t.Fatalf("expected AnzahlZuendungen 1, got %d", kesselRecord.AnzahlZuendungen.Value)
	}

	// Start Leistungsbrand (Zündung endet)
	// 14:10:40 bis 14:20:20 sind 9 Minuten und 40 Sekunden = 540 + 40 = 580 Sekunden
	handleZRecord([]string{"z", "14:20:20", "Kessel", "Leistungsbrand"}, "z 14:20:20 Kessel Leistungsbrand")
	if kesselRecord.DauerLetzteZuendung.Value != 580 {
		t.Fatalf("expected DauerLetzteZuendung 580, got %d", kesselRecord.DauerLetzteZuendung.Value)
	}

	// Test mit anderer Kodierung/Schreibweise, die durch Z.*ndung abgedeckt sein sollte
	handleZRecord([]string{"z", "14:30:00", "Kessel", "Zündungen"}, "z 14:30:00 Kessel Zündungen")
	if kesselRecord.AnzahlZuendungen.Value != 2 {
		t.Fatalf("expected AnzahlZuendungen 2 after Zündungen, got %d", kesselRecord.AnzahlZuendungen.Value)
	}
	handleZRecord([]string{"z", "14:35:00", "Kessel", "Leistungsbrand"}, "z 14:35:00 Kessel Leistungsbrand")
	// 14:30:00 bis 14:35:00 sind 5 Minuten = 300 Sekunden
	if kesselRecord.DauerLetzteZuendung.Value != 300 {
		t.Fatalf("expected DauerLetzteZuendung 300, got %d", kesselRecord.DauerLetzteZuendung.Value)
	}

	// Noch ein Test für Zndung
	handleZRecord([]string{"z", "14:40:00", "Kessel", "Zndung"}, "z 14:40:00 Kessel Zndung")
	if kesselRecord.AnzahlZuendungen.Value != 3 {
		t.Fatalf("expected AnzahlZuendungen 3 after Zndung, got %d", kesselRecord.AnzahlZuendungen.Value)
	}
	handleZRecord([]string{"z", "14:45:00", "Kessel", "Leistungsbrand"}, "z 14:45:00 Kessel Leistungsbrand")
	if kesselRecord.DauerLetzteZuendung.Value != 300 {
		t.Fatalf("expected DauerLetzteZuendung 300 (second time), got %d", kesselRecord.DauerLetzteZuendung.Value)
	}

	// Ende Leistungsbrand
	// Von 14:45:00 bis 18:00:32
	// 14:45:00 -> 17:45:00 sind 3 Stunden = 10800s
	// 17:45:00 -> 18:00:00 sind 15 Minuten = 900s
	// 18:00:00 -> 18:00:32 sind 32s
	// Gesamt: 10800 + 900 + 32 = 11732
	handleZRecord([]string{"z", "18:00:32", "Kessel", "Aus"}, "z 18:00:32 Kessel Aus")
	if kesselRecord.DauerLetzterLeistungsbrand.Value != 11732 {
		t.Fatalf("expected DauerLetzterLeistungsbrand 11732, got %d", kesselRecord.DauerLetzterLeistungsbrand.Value)
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
