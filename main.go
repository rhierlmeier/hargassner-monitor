package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/creativeprojects/go-homie"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.bug.st/serial"
)

var mqttClient mqtt.Client
var homieDevice = homie.NewDevice("hargassner", "Hargassner Heizung")
var nodeProcessWerte = homieDevice.AddNode("prozesswerte", "Prozesswerte", "Prozesswerte")
var nodeHeizkreis1 = homieDevice.AddNode("heizkreis1", "Heizkreis 1", "Heizkreis 1")
var nodeHeizkreis2 = homieDevice.AddNode("heizkreis2", "Heizkreis 2", "Heizkreis 2")
var nodeStoerung = homieDevice.AddNode("stoerung", "Störung", "Störung")

var stoerungRecord = newEmptyStoerungRecord(nodeStoerung)

var meldung = newMeldung(nodeProcessWerte)

type MultiLanguageString struct {
	EN string
	DE string
}

func newMeldung(node *homie.Node) StatusField[string] {
	ret := StatusField[string]{
		Id:   "meldung",
		Name: MultiLanguageString{EN: "Message", DE: "Meldung"},
		Unit: "",
	}
	registerStatusField(&ret, node)
	return ret
}

type StatusField[T any] struct {
	Id            string
	Value         T
	Name          MultiLanguageString
	Unit          string
	HomieProperty *homie.Property
}

type StatusRecord struct {
	PrimaryAirFan              StatusField[int]
	ExhaustFan                 StatusField[int]
	O2InExhaustGas             StatusField[float64]
	BoilerTemperature          StatusField[int]
	ExhaustGasTemperature      StatusField[int]
	CurrentOutdoorTemperature  StatusField[float64]
	AverageOutdoorTemperature  StatusField[float64]
	FlowTemperatureCircuit1    StatusField[float64]
	FlowTemperatureCircuit2    StatusField[float64]
	FlowTemperatureCircuit1Set StatusField[float64]
	FlowTemperatureCircuit2Set StatusField[float64]
	ReturnBoiler2BufferTemp    StatusField[int]
	BoilerTemperature1         StatusField[int]
	FeedRate                   StatusField[int]
	BoilerSetTemperature       StatusField[int]
	CurrentUnderpressure       StatusField[float64]
	AverageUnderpressure       StatusField[float64]
	SetUnderpressure           StatusField[float64]
	BoilerTemperature2SM       StatusField[float64]
	HK1FR25                    StatusField[float64]
	HK2FR25                    StatusField[float64]
	MotorCurrentFeedScrew      StatusField[float64]
	MotorCurrentAshDischarge   StatusField[float64]
	MotorCurrentRoomDischarge  StatusField[float64]
}

type StoerungRecord struct {
	StoerungNr     StatusField[int]
	StoerungText   StatusField[string]
	StoerungActive StatusField[bool]
	LastActive     StatusField[string]
}

func newEmptyStoerungRecord(node *homie.Node) *StoerungRecord {
	ret := &StoerungRecord{
		StoerungNr:     StatusField[int]{Id: "nr", Name: MultiLanguageString{EN: "Error Number", DE: "Störungsnummer"}, Unit: ""},
		StoerungText:   StatusField[string]{Id: "text", Name: MultiLanguageString{EN: "Error Text", DE: "Störungstext"}, Unit: ""},
		StoerungActive: StatusField[bool]{Id: "active", Name: MultiLanguageString{EN: "Error Active", DE: "Störung Aktiv"}, Unit: ""},
		LastActive:     StatusField[string]{Id: "lastActive", Name: MultiLanguageString{EN: "Last Active", DE: "Letzte Aktivität"}, Unit: ""},
	}

	registerStatusField(&ret.StoerungNr, node)
	registerStatusField(&ret.StoerungText, node)
	registerStatusField(&ret.StoerungActive, node)
	registerStatusField(&ret.LastActive, node)

	return ret
}

func (field *StatusField[T]) SetValue(value T) {
	field.Value = value
	if field.HomieProperty != nil {
		field.HomieProperty.Set(value)
	}
}

func newEmptyStatusRecord() *StatusRecord {
	return &StatusRecord{
		PrimaryAirFan:              StatusField[int]{Id: "primaerLuftGeblaese", Name: MultiLanguageString{EN: "Primary Air Fan", DE: "Primärluftgebläse"}, Unit: "%"},
		ExhaustFan:                 StatusField[int]{Id: "saugluftGeblaese", Name: MultiLanguageString{EN: "Exhaust Fan", DE: "Saugluftgebläse"}, Unit: "%"},
		O2InExhaustGas:             StatusField[float64]{Id: "o2InAbgas", Name: MultiLanguageString{EN: "O2 in Exhaust Gas", DE: "O2 im Abgas"}, Unit: "%"},
		BoilerTemperature:          StatusField[int]{Id: "kesselTemperatur", Name: MultiLanguageString{EN: "Boiler Temperature", DE: "Kesseltemperatur"}, Unit: "°C"},
		ExhaustGasTemperature:      StatusField[int]{Id: "rauchgasTemperatur", Name: MultiLanguageString{EN: "Exhaust Gas Temperature", DE: "Rauchgastemperatur"}, Unit: "°C"},
		CurrentOutdoorTemperature:  StatusField[float64]{Id: "aussenTemperaturAktuell", Name: MultiLanguageString{EN: "Current Outdoor Temperature", DE: "Außentemperatur aktuell"}, Unit: "°C"},
		AverageOutdoorTemperature:  StatusField[float64]{Id: "aussenTemperaturGemittelt", Name: MultiLanguageString{EN: "Average Outdoor Temperature", DE: "Außentemperatur gemittelt"}, Unit: "°C"},
		FlowTemperatureCircuit1:    StatusField[float64]{Id: "vorlaufTemperatur", Name: MultiLanguageString{EN: "Flow Temperature Circuit 1", DE: "Vorlauftemperatur Kreis 1"}, Unit: "°C"},
		FlowTemperatureCircuit2:    StatusField[float64]{Id: "vorlaufTemperatur", Name: MultiLanguageString{EN: "Flow Temperature Circuit 2", DE: "Vorlauftemperatur Kreis 2"}, Unit: "°C"},
		FlowTemperatureCircuit1Set: StatusField[float64]{Id: "vorlaufSollTemperatur", Name: MultiLanguageString{EN: "Flow Temperature Circuit 1 Set", DE: "Soll-Vorlauftemperatur Kreis 1"}, Unit: "°C"},
		FlowTemperatureCircuit2Set: StatusField[float64]{Id: "vorlaufSollTemperatur", Name: MultiLanguageString{EN: "Flow Temperature Circuit 2 Set", DE: "Soll-Vorlauftemperatur Kreis 2"}, Unit: "°C"},
		ReturnBoiler2BufferTemp:    StatusField[int]{Id: "ReturnBoiler2BufferTemp", Name: MultiLanguageString{EN: "Return Boiler to Buffer Temperature", DE: "Rücklauf Kessel zu Puffer Temperatur"}, Unit: "°C"},
		BoilerTemperature1:         StatusField[int]{Id: "boiler1Temperatur", Name: MultiLanguageString{EN: "Boiler Temperature 1", DE: "Kesseltemperatur 1"}, Unit: "°C"},
		FeedRate:                   StatusField[int]{Id: "foerderMenge", Name: MultiLanguageString{EN: "Feed Rate", DE: "Fördermenge"}, Unit: "%"},
		BoilerSetTemperature:       StatusField[int]{Id: "boiler1SollTemperatur", Name: MultiLanguageString{EN: "Boiler1 Set Temperature", DE: "Solltemperatur Boiler1"}, Unit: "°C"},
		CurrentUnderpressure:       StatusField[float64]{Id: "unterdruckAktuell", Name: MultiLanguageString{EN: "Current Underpressure", DE: "Unterdruck aktuell"}, Unit: "Pa"},
		AverageUnderpressure:       StatusField[float64]{Id: "unterdruckGemittelt", Name: MultiLanguageString{EN: "Average Underpressure", DE: "Unterdruck gemittelt"}, Unit: "Pa"},
		SetUnderpressure:           StatusField[float64]{Id: "unterdruckSoll", Name: MultiLanguageString{EN: "Set Underpressure", DE: "Soll-Unterdruck"}, Unit: "Pa"},
		BoilerTemperature2SM:       StatusField[float64]{Id: "BoilerTemperature2SM", Name: MultiLanguageString{EN: "Boiler Temperature 2 SM", DE: "Kesseltemperatur 2 SM"}, Unit: "°C"},
		HK1FR25:                    StatusField[float64]{Id: "HK1FR25", Name: MultiLanguageString{EN: "HK1 FR25", DE: "HK1 FR25"}, Unit: "°C"},
		HK2FR25:                    StatusField[float64]{Id: "HK2FR25", Name: MultiLanguageString{EN: "HK2 FR25", DE: "HK2 FR25"}, Unit: "°C"},
		MotorCurrentFeedScrew:      StatusField[float64]{Id: "stromEinschub", Name: MultiLanguageString{EN: "Motor Current Feed Screw", DE: "Strom Einschub"}, Unit: "A"},
		MotorCurrentAshDischarge:   StatusField[float64]{Id: "stromAscheaustragung", Name: MultiLanguageString{EN: "Motor Current Ash Discharge", DE: "Strom Ascheaustragung"}, Unit: "A"},
		MotorCurrentRoomDischarge:  StatusField[float64]{Id: "stromRaumaustragung", Name: MultiLanguageString{EN: "Motor Current Room Discharge", DE: "Strom Raumaustragung"}, Unit: "A"},
	}

}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is ready"))
}

func parseField[T any](fields []string, index int, field StatusField[T]) {
	if index >= len(fields) {
		log.Fatalf("index %d out of range for fields", index)
	}
	value := fields[index]
	var fieldValue T
	switch any(field.Value).(type) {
	case int:
		parsedValue, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("invalid int field[%d] [%s]: %v (fields: %s)", index, field.Id, err, strings.Join(fields, "|"))
			return
		}
		fieldValue = any(parsedValue).(T)
	case float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("invalid float field[%d] [%s]: %v (fields: %s)", index, field.Id, err, strings.Join(fields, "|"))
			return
		}
		fieldValue = any(parsedValue).(T)
	case string:
		fieldValue = any(value).(T)
	default:
		log.Fatalf("unsupported consumer type for field %s at %d", field.Id, index)
	}
	field.SetValue(fieldValue)
}

func parseStatusRecord(fields []string, record *StatusRecord) error {
	if len(fields) < 32 {
		return fmt.Errorf("not enough fields")
	}

	parseField(fields, 1, record.PrimaryAirFan)
	parseField(fields, 2, record.ExhaustFan)
	parseField(fields, 3, record.O2InExhaustGas)
	parseField(fields, 4, record.BoilerTemperature)
	parseField(fields, 5, record.ExhaustGasTemperature)
	parseField(fields, 6, record.CurrentOutdoorTemperature)
	parseField(fields, 7, record.AverageOutdoorTemperature)
	parseField(fields, 8, record.FlowTemperatureCircuit1)
	parseField(fields, 9, record.FlowTemperatureCircuit2)
	parseField(fields, 10, record.FlowTemperatureCircuit1Set)
	parseField(fields, 11, record.FlowTemperatureCircuit2Set)
	parseField(fields, 12, record.ReturnBoiler2BufferTemp)
	parseField(fields, 13, record.BoilerTemperature1)
	parseField(fields, 14, record.FeedRate)
	parseField(fields, 15, record.BoilerSetTemperature)
	parseField(fields, 16, record.CurrentUnderpressure)
	parseField(fields, 17, record.AverageUnderpressure)
	parseField(fields, 18, record.SetUnderpressure)
	// Field 19 to 23 are for Heizkreis 3 and 4
	parseField(fields, 23, record.BoilerTemperature2SM)
	parseField(fields, 24, record.HK1FR25)
	parseField(fields, 25, record.HK2FR25)

	// Field 26 and 27 for Heizkreis 3 and 4

	// 28, 29, 30, 31 are not used
	parseField(fields, 29, record.MotorCurrentFeedScrew)
	parseField(fields, 30, record.MotorCurrentAshDischarge)
	parseField(fields, 31, record.MotorCurrentRoomDischarge)

	return nil
}

func getEnv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// topicToValue is a map of topics to values to avoid sending the same value multiple times
var topicToValue = make(map[string]string)

// onSet handles the setting of a topic's value and publishes the value if it has changed.
// It ensures that blank strings are not sent for non-string data types.
//
// Parameters:
//   - topic: The topic whose value is being set.
//   - value: The new value to set for the topic.
//   - dataType: The data type of the property being set.
//
// Behavior:
//   - If the value is "<nil>", it is converted to an empty string.
//   - If the value is an empty string and the data type is not a string, the function returns without publishing.
//   - If the value has changed from the previous value for the topic, it updates the value and publishes it.
func onSet(topic, value string, dataType homie.PropertyType) {
	if value == "<nil>" {
		value = ""
	}
	if value == "" && dataType != homie.TypeString {
		// don't send a blank string on anything else than a string data type
		return
	}
	// Publish the values only when it really changes
	if topicToValue[topic] != value {
		topicToValue[topic] = value
		publish(topic, value)
	}
}

func publish(topic, value string) {
	mqttClient.Publish(topic, 0, false, value)
}

func onConnectionLost(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
	for k := range topicToValue {
		delete(topicToValue, k)
	}
}

func onConnected(client mqtt.Client) {
	log.Printf("Connected to MQTT broker")
	publishAllHomieAttributes()
}

func publishAllHomieAttributes() {
	// get the full homie definition to send to MQTT - you only need to send it once unless it's changing over time
	for _, attribute := range homieDevice.GetHomieAttributes() {
		mqttClient.Publish(attribute.Topic, 0, true, attribute.Value)
	}
}

func getStoerungText(stoerNr int) string {
	stoerungText := map[int]string{
		1:  "Sicherung F25 defekt",
		2:  "Elektronischer Motorschutz Einschubschnecke ausgelöst",
		3:  "Elektronischer Motorschutz Raumaustragung ausgelöst",
		4:  "Elektronischer Motorschutz Ascheaustragung ausgelöst",
		5:  "Sicherheitsthermostat (STB)",
		6:  "Rücklaufzeit überschritten",
		7:  "Endschalter Deckel offen",
		8:  "Brennraum überfüllt",
		9:  "Brandschutzklappe öffnet nicht",
		10: "Zündzeit überschritten",
		11: "Minimale Rauchgastemperatur unterschritten",
		12: "Initiator Entaschung",
		13: "Überstrom Einschubschnecke",
		14: "Überstrom Raumaustragung",
		15: "Überstrom Aschenaustragung",
		16: "Rauchgasfühler falsch angeschlossen",
		17: "Rauchgasfühler Unterbrechung",
		18: "Kesselfühler Kurzschluss",
		19: "Kesselfühler Unterbrechung",
		20: "Boilerfühler 1 Kurzschluss",
	}

	text, ok := stoerungText[stoerNr]
	if !ok {
		text = "Unbekannte Störung"
	}
	return text
}

func registerStatusField[T any](field *StatusField[T], node *homie.Node) {

	var propertyType homie.PropertyType

	switch any(field.Value).(type) {
	case int:
		propertyType = homie.TypeInteger
	case float64:
		propertyType = homie.TypeFloat
	case bool:
		propertyType = homie.TypeBoolean
	case string:
		propertyType = homie.TypeString
	default:
		log.Fatalf("unsupported type of field %s", field.Id)
	}
	field.HomieProperty = node.AddProperty(field.Id, field.Name.EN, propertyType).SetUnit(field.Unit)
}

func main() {
	serialDevice := getEnv("HARGASSNER_SERIAL_DEVICE", "/dev/ttyUSB0")

	mode := &serial.Mode{
		BaudRate: 19200,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(serialDevice, mode)
	if err != nil {
		log.Fatalf("could not open %s: %s", serialDevice, err)
	}

	statusRecord := newEmptyStatusRecord()

	registerStatusField(&statusRecord.PrimaryAirFan, nodeProcessWerte)
	registerStatusField(&statusRecord.ExhaustFan, nodeProcessWerte)
	registerStatusField(&statusRecord.O2InExhaustGas, nodeProcessWerte)
	registerStatusField(&statusRecord.BoilerTemperature, nodeProcessWerte)
	registerStatusField(&statusRecord.ExhaustGasTemperature, nodeProcessWerte)
	registerStatusField(&statusRecord.CurrentOutdoorTemperature, nodeProcessWerte)
	registerStatusField(&statusRecord.AverageOutdoorTemperature, nodeProcessWerte)
	registerStatusField(&statusRecord.FlowTemperatureCircuit1, nodeHeizkreis1)
	registerStatusField(&statusRecord.FlowTemperatureCircuit2, nodeHeizkreis2)
	registerStatusField(&statusRecord.FlowTemperatureCircuit1Set, nodeHeizkreis1)
	registerStatusField(&statusRecord.FlowTemperatureCircuit2Set, nodeHeizkreis2)
	registerStatusField(&statusRecord.ReturnBoiler2BufferTemp, nodeProcessWerte)
	registerStatusField(&statusRecord.BoilerTemperature1, nodeProcessWerte)
	registerStatusField(&statusRecord.FeedRate, nodeProcessWerte)
	registerStatusField(&statusRecord.BoilerSetTemperature, nodeProcessWerte)
	registerStatusField(&statusRecord.CurrentUnderpressure, nodeProcessWerte)
	registerStatusField(&statusRecord.AverageUnderpressure, nodeProcessWerte)
	registerStatusField(&statusRecord.SetUnderpressure, nodeProcessWerte)
	registerStatusField(&statusRecord.BoilerTemperature2SM, nodeProcessWerte)
	registerStatusField(&statusRecord.HK1FR25, nodeHeizkreis1)
	registerStatusField(&statusRecord.HK2FR25, nodeHeizkreis2)
	registerStatusField(&statusRecord.MotorCurrentFeedScrew, nodeProcessWerte)
	registerStatusField(&statusRecord.MotorCurrentAshDischarge, nodeProcessWerte)
	registerStatusField(&statusRecord.MotorCurrentRoomDischarge, nodeProcessWerte)
	registerStatusField(&statusRecord.MotorCurrentRoomDischarge, nodeProcessWerte)

	homieDevice.OnSet(onSet)

	httpPort := getEnv("HARGASSNER_MONITOR_PORT", "8080")

	http.HandleFunc("/readiness", readinessProbe)

	go func() {
		log.Fatal(http.ListenAndServe(":"+httpPort, nil))
	}()

	opts := mqtt.NewClientOptions().
		AddBroker(getEnv("HARGASSNER_MQTT_BROKER", "tcp://localhost:1883")).
		SetClientID(getEnv("HARGASSNER_MQTT_CLIENT_ID", "hargassner-monitor")).
		SetUsername(getEnv("HARGASSNER_MQTT_USER", "")).
		SetPassword(getEnv("HARGASSNER_MQTT_PASSWORD", ""))

	opts.SetAutoReconnect(true)

	opts.OnConnectionLost = onConnectionLost
	opts.OnConnect = onConnected

	log.Printf("Connecting to MQTT broker %s", opts.Servers[0])

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	publishAllHomieAttributes()

	log.Printf("Reading from on %s", serialDevice)
	reader := bufio.NewReader(port)

	reader.ReadString('\n')

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		line, err = strconv.Unquote(strings.Replace(strconv.Quote(line), `\\x`, `\x`, -1))
		if err != nil {
			log.Fatal(err)
		}

		fields := strings.Fields(strings.TrimSpace(line))

		log.Printf("Received fields: %s", strings.Join(fields, "|"))

		switch fields[0] {
		case "pm":
			err := parseStatusRecord(fields, statusRecord)
			if err != nil {
				log.Println("Error parsing status record:", err)
				continue
			}
		case "z":
			{
				handleZRecord(fields, line)
			}
		default:
			fmt.Print("Unknown record receive:" + line)
		}
	}

}

func handleZRecord(fields []string, line string) {

	log.Printf("Handling Z record: fields:[%s]", strings.Join(fields, "|"))
	isStoerung := strings.HasPrefix(fields[2], "St") && strings.HasSuffix(fields[2], "rung")
	if isStoerung {
		// 0.1........2........3...4.5
		// z 18:39:41 Stoerung Set 7 Stop:1
		// 0.1........2.......3....4
		// z 18:40:16 Störung Quit 0007

		var active bool
		switch fields[3] {
		case "Set":
			active = true
		case "Quit":
			active = false
		default:
			log.Printf("Unexpected value of fields[3] (%s): %s (line: %s)", fields[3], "Set or Quit", line)
			return
		}

		stoerNr, err := strconv.Atoi(fields[4])
		if err != nil {
			log.Printf("Unexpected value of fields[4] (%s): %s (line: %s)", fields[4], err, line)
			return
		}
		stoerungText := getStoerungText(stoerNr)

		if active {
			log.Printf("Störung %d: %s", stoerNr, stoerungText)
		} else {
			log.Printf("Quit Störung %d: %s", stoerNr, stoerungText)
		}

		lastChange := fields[1]

		stoerungRecord.StoerungNr.SetValue(stoerNr)
		stoerungRecord.StoerungText.SetValue(stoerungText)
		stoerungRecord.StoerungActive.SetValue(active)
		stoerungRecord.LastActive.SetValue(lastChange)

	} else {
		message := strings.Join(fields[2:], " ")
		meldung.SetValue(message)
	}
}
