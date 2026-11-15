package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/go-homie"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsEndpoint(t *testing.T) {
	// Re-register or use existing handlers if possible
	// Since main uses http.Handle on default mux, we might have issues if already registered.
	// But in test we can create a new mux.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check if it contains some prometheus default metrics at least
	if !strings.Contains(rr.Body.String(), "go_info") {
		t.Errorf("handler returned unexpected body: %v", rr.Body.String())
	}
}

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
	// LastActive carries the time field from input (fields[1]) as timestamp
	if !strings.Contains(stoerungRecord.LastActive.Value, "T18:40:16") {
		t.Fatalf("after Quit expected LastActive at 18:40:16, got %q", stoerungRecord.LastActive.Value)
	}
	if _, err := time.Parse(time.RFC3339, stoerungRecord.LastActive.Value); err != nil {
		t.Fatalf("LastActive is no RFC 3339 timestamp: %v", err)
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

func TestReadinessProblems(t *testing.T) {
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	oldMax := maxDataAge
	maxDataAge = 5 * time.Minute
	t.Cleanup(func() { maxDataAge = oldMax })
	lastSerialData.Store(start.UnixNano())

	if p := readinessProblems(start.Add(5*time.Minute), true); len(p) != 0 {
		t.Errorf("within max data age: %v", p)
	}

	p := readinessProblems(start.Add(6*time.Minute), true)
	if len(p) != 1 || !strings.Contains(p[0], "no data from serial port for 6m0s") {
		t.Errorf("stale data: %v", p)
	}

	p = readinessProblems(start.Add(time.Minute), false)
	if len(p) != 1 || p[0] != "mqtt not connected" {
		t.Errorf("mqtt disconnected: %v", p)
	}

	if p := readinessProblems(start.Add(10*time.Minute), false); len(p) != 2 {
		t.Errorf("both problems expected: %v", p)
	}
}

func TestReadinessProbeHandler(t *testing.T) {
	// mqttClient is nil in tests, so the probe must report not ready.
	lastSerialData.Store(time.Now().UnixNano())
	rr := httptest.NewRecorder()
	readinessProbe(rr, httptest.NewRequest("GET", "/readiness", nil))
	if rr.Code != http.StatusServiceUnavailable || !strings.Contains(rr.Body.String(), "mqtt not connected") {
		t.Errorf("got %d %q", rr.Code, rr.Body.String())
	}
}

func TestLastSerialDataSeconds(t *testing.T) {
	ts := time.Date(2026, 11, 12, 6, 0, 0, 500_000_000, time.UTC)
	lastSerialData.Store(ts.UnixNano())
	if got, want := lastSerialDataSeconds(), float64(ts.Unix())+0.5; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLatin1ToUTF8(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Kessel Z\xfcndung", "Kessel Zündung"},
		{"RLM Restw\xe4rme", "RLM Restwärme"},
		{"z 18:40:16 St\xf6rung Quit 0007", "z 18:40:16 Störung Quit 0007"},
		{"pm 100 7.8 73", "pm 100 7.8 73"},
		{"Kessel Zündung", "Kessel Zündung"},
	}
	for _, tt := range tests {
		if got := latin1ToUTF8(tt.in); got != tt.want {
			t.Errorf("latin1ToUTF8(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestControllerTimestamp(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 15, 8, 30, 0, 0, loc)
	cases := []struct {
		clock string
		want  string
	}{
		{"08:29:10", "2026-11-15T08:29:10+01:00"},
		// slightly ahead (controller clock runs fast): still today
		{"09:00:00", "2026-11-15T09:00:00+01:00"},
		// more than an hour ahead: record from before midnight
		{"23:59:30", "2026-11-14T23:59:30+01:00"},
	}
	for _, c := range cases {
		if got := controllerTimestamp(c.clock, now); got != c.want {
			t.Errorf("controllerTimestamp(%q) = %q, want %q", c.clock, got, c.want)
		}
	}
	// invalid time: falls back to now
	if got := controllerTimestamp("xx", now); got != "2026-11-15T08:30:00+01:00" {
		t.Errorf("controllerTimestamp(invalid) = %q", got)
	}
}

// registerAllForTest registers every property like main does.
func registerAllForTest() {
	registerStatusRecord(newEmptyStatusRecord())
}

func TestHADiscoveryEntities(t *testing.T) {
	registerAllForTest()

	// 24 process values, 3 Kessel, 4 Störung, 1 Meldung
	if len(haEntities) != 32 {
		t.Fatalf("expected 32 properties, got %d", len(haEntities))
	}
	for key := range haOverrides {
		if _, ok := haEntities[key]; !ok {
			t.Errorf("override for unknown property %s", key)
		}
	}

	uniqueIDs := map[string]string{}
	topics := map[string]string{}
	legacy := 0
	for key, e := range haEntities {
		c := e.config("en")
		if other, dup := uniqueIDs[c.UniqueID]; dup {
			t.Errorf("unique_id %s used by %s and %s", c.UniqueID, key, other)
		}
		uniqueIDs[c.UniqueID] = key
		topic := e.configTopic("homeassistant")
		if other, dup := topics[topic]; dup {
			t.Errorf("config topic %s used by %s and %s", topic, key, other)
		}
		topics[topic] = key

		if c.StateTopic != "homie/hargassner/"+key {
			t.Errorf("%s: state_topic %s", key, c.StateTopic)
		}
		if haOverrides[key].LegacyUniqueID != "" {
			legacy++
			if c.DefaultEntityID != "" {
				t.Errorf("%s: legacy entity must not set default_entity_id", key)
			}
		} else if c.DefaultEntityID != e.component()+"."+c.UniqueID {
			t.Errorf("%s: default_entity_id %q", key, c.DefaultEntityID)
		}
		wantComponent := "sensor"
		if e.DataType == homie.TypeBoolean {
			wantComponent = "binary_sensor"
		}
		if !strings.HasPrefix(topic, "homeassistant/"+wantComponent+"/hargassner/") {
			t.Errorf("%s: config topic %s", key, topic)
		}
	}
	// the entities of the former hand-written Home Assistant config
	if legacy != 23 {
		t.Errorf("expected 23 legacy unique_ids, got %d", legacy)
	}
}

func TestHADiscoveryConfig(t *testing.T) {
	registerAllForTest()

	c := haEntities["prozesswerte/aussenTemperaturAktuell"].config("de")
	if c.Name != "Außentemperatur aktuell" || c.UniqueID != "hargassner_aussentemperatur_aktuell" ||
		c.UnitOfMeasurement != "°C" || c.DeviceClass != "temperature" || c.StateClass != "measurement" {
		t.Errorf("aussenTemperaturAktuell: %+v", c)
	}
	if c.Availability[0].Topic != "homie/hargassner/$state" {
		t.Errorf("availability topic %s", c.Availability[0].Topic)
	}

	c = haEntities["stoerung/active"].config("en")
	if c.PayloadOn != "true" || c.PayloadOff != "false" || c.DeviceClass != "problem" || c.UnitOfMeasurement != "" {
		t.Errorf("stoerung/active: %+v", c)
	}

	// a code without unit gets no statistics
	if c = haEntities["stoerung/nr"].config("en"); c.StateClass != "" {
		t.Errorf("stoerung/nr: state_class %q", c.StateClass)
	}
	if c = haEntities["kessel/AnzahlZuendungen"].config("en"); c.StateClass != "total_increasing" {
		t.Errorf("AnzahlZuendungen: state_class %q", c.StateClass)
	}

	c = haEntities["stoerung/text"].config("en")
	if c.UniqueID != "hargassner_stoerung_text" || c.DefaultEntityID != "sensor.hargassner_stoerung_text" {
		t.Errorf("stoerung/text: %+v", c)
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "unit_of_measurement") || strings.Contains(string(data), "payload_on") {
		t.Errorf("empty fields not omitted: %s", data)
	}
}

func TestParseHADiscovery(t *testing.T) {
	registerAllForTest()

	if selected, err := parseHADiscovery(""); selected != nil || err != nil {
		t.Errorf("empty: %v %v", selected, err)
	}
	if selected, err := parseHADiscovery("all"); err != nil || len(selected) != len(haEntities) {
		t.Errorf("all: %d %v", len(selected), err)
	}
	selected, err := parseHADiscovery("prozesswerte/aussenTemperaturAktuell, stoerung/active")
	if err != nil || len(selected) != 2 || !selected["stoerung/active"] {
		t.Errorf("list: %v %v", selected, err)
	}
	if _, err := parseHADiscovery("prozesswerte/gibtsNicht"); err == nil {
		t.Error("unknown property accepted")
	}
}

func TestHADiscoveryMessages(t *testing.T) {
	registerAllForTest()

	selected, _ := parseHADiscovery("prozesswerte/aussenTemperaturAktuell")
	messages, err := haDiscoveryMessages(selected, "homeassistant", "de")
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != len(haEntities) {
		t.Fatalf("expected one message per property, got %d", len(messages))
	}
	payload := messages["homeassistant/sensor/hargassner/prozesswerte_aussenTemperaturAktuell/config"]
	var c haConfig
	if err := json.Unmarshal([]byte(payload), &c); err != nil || c.UniqueID != "hargassner_aussentemperatur_aktuell" {
		t.Errorf("selected property: %q %v", payload, err)
	}
	// not selected: empty payload removes an earlier discovered entity
	if payload, ok := messages["homeassistant/sensor/hargassner/prozesswerte_kesselTemperatur/config"]; !ok || payload != "" {
		t.Errorf("unselected property: %q %v", payload, ok)
	}
}
