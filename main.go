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
var homieDevice *homie.Device

const (
	DEVICE_NAME = "hargassner"

	NODE_PROCESSWERTE = "prozesswerte"
	NODE_HEIZKREIS1   = "heizkreis1"
	NODE_HEIZKREIS2   = "heizkreis2"
	NODE_STOERUNG     = "stoerung"

	PROPERTY_RAUCHGAS_TEMPERATUR         = "rauchgasTemperatur"
	PROPERTY_BOILER1_TEMPERATUR          = "boiler1Temperatur"
	PROPERTY_AUSSEN_TEMPERATUR_AKTUELL   = "aussenTemperaturAktuell"
	PROPERTY_AUSSEN_TEMPERATUR_GEMITTELT = "aussenTemperaturGemittelt"
	PROPERTY_KESSEL_TEMPERATUR           = "kesselTemperatur"
	PROPERTY_KESSEL_SOLL_TEMPERATUR      = "kesselSollTemperatur"
	PROPERTY_MELDUNG                     = "meldung"
	PROPERTY_SAUGLUFT_GEBLÄSE            = "saugluftGeblaese"
	PROPERTY_PRIMÄR_LUFT_GEBLÄSE         = "primaerLuftGeblaese"
	PROPERTY_O2_IN_ABGAS                 = "o2InAbgas"
	PROPERTY_FÖRDER_MENGE                = "foerderMenge"
	PROPERTY_STROM_RAUMAUSTRAGUNG        = "stromRaumaustragung"
	PROPERTY_STROM_ASCHEAUSTRAGUNG       = "stromAscheaustragung"
	PROPERTY_STROM_EINSCHUB              = "stromEinschub"
	PROPERTY_VORLAUF_TEMPERATUR          = "vorlaufTemperatur"
	PROPERTY_VORLAUF_SOLL_TEMPERATUR     = "vorlaufSollTemperatur"

	PROPERTY_NR          = "nr"
	PROPERTY_TEXT        = "text"
	PROPERTY_ACTIVE      = "active"
	PROPERTY_LAST_CHANGE = "lastChange"
)

type MultiLanguageString struct {
	EN string
	DE string
}

type StatusField[T any] struct {
	Id    string
	Value T
	Name  MultiLanguageString
	Unit  string
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
	FlowTemperatureCircuit3    StatusField[float64]
	FlowTemperatureCircuit4    StatusField[float64]
	FlowTemperatureCircuit3Set StatusField[float64]
	FlowTemperatureCircuit4Set StatusField[float64]
	BoilerTemperature2SM       StatusField[float64]
	HK1FR25                    StatusField[float64]
	HK2FR25                    StatusField[float64]
	HK3FR25SM                  StatusField[float64]
	HK4FR25SM                  StatusField[float64]
	MotorCurrentFeedScrew      StatusField[float64]
	MotorCurrentAshDischarge   StatusField[float64]
	MotorCurrentRoomDischarge  StatusField[float64]
}

func newEmptyStatusRecord() *StatusRecord {
	return &StatusRecord{
		PrimaryAirFan:              StatusField[int]{Id: "PrimaryAirFan", Name: MultiLanguageString{EN: "Primary Air Fan", DE: "Primärluftgebläse"}, Unit: "%"},
		ExhaustFan:                 StatusField[int]{Id: "ExhaustFan", Name: MultiLanguageString{EN: "Exhaust Fan", DE: "Abgasgebläse"}, Unit: "%"},
		O2InExhaustGas:             StatusField[float64]{Id: "O2InExhaustGas", Name: MultiLanguageString{EN: "O2 in Exhaust Gas", DE: "O2 im Abgas"}, Unit: "%"},
		BoilerTemperature:          StatusField[int]{Id: "BoilerTemperature", Name: MultiLanguageString{EN: "Boiler Temperature", DE: "Kesseltemperatur"}, Unit: "°C"},
		ExhaustGasTemperature:      StatusField[int]{Id: "ExhaustGasTemperature", Name: MultiLanguageString{EN: "Exhaust Gas Temperature", DE: "Rauchgastemperatur"}, Unit: "°C"},
		CurrentOutdoorTemperature:  StatusField[float64]{Id: "CurrentOutdoorTemperature", Name: MultiLanguageString{EN: "Current Outdoor Temperature", DE: "Aktuelle Außentemperatur"}, Unit: "°C"},
		AverageOutdoorTemperature:  StatusField[float64]{Id: "AverageOutdoorTemperature", Name: MultiLanguageString{EN: "Average Outdoor Temperature", DE: "Durchschnittliche Außentemperatur"}, Unit: "°C"},
		FlowTemperatureCircuit1:    StatusField[float64]{Id: "FlowTemperatureCircuit1", Name: MultiLanguageString{EN: "Flow Temperature Circuit 1", DE: "Vorlauftemperatur Kreis 1"}, Unit: "°C"},
		FlowTemperatureCircuit2:    StatusField[float64]{Id: "FlowTemperatureCircuit2", Name: MultiLanguageString{EN: "Flow Temperature Circuit 2", DE: "Vorlauftemperatur Kreis 2"}, Unit: "°C"},
		FlowTemperatureCircuit1Set: StatusField[float64]{Id: "FlowTemperatureCircuit1Set", Name: MultiLanguageString{EN: "Flow Temperature Circuit 1 Set", DE: "Soll-Vorlauftemperatur Kreis 1"}, Unit: "°C"},
		FlowTemperatureCircuit2Set: StatusField[float64]{Id: "FlowTemperatureCircuit2Set", Name: MultiLanguageString{EN: "Flow Temperature Circuit 2 Set", DE: "Soll-Vorlauftemperatur Kreis 2"}, Unit: "°C"},
		ReturnBoiler2BufferTemp:    StatusField[int]{Id: "ReturnBoiler2BufferTemp", Name: MultiLanguageString{EN: "Return Boiler to Buffer Temperature", DE: "Rücklauf Kessel zu Puffer Temperatur"}, Unit: "°C"},
		BoilerTemperature1:         StatusField[int]{Id: "BoilerTemperature1", Name: MultiLanguageString{EN: "Boiler Temperature 1", DE: "Kesseltemperatur 1"}, Unit: "°C"},
		FeedRate:                   StatusField[int]{Id: "FeedRate", Name: MultiLanguageString{EN: "Feed Rate", DE: "Fördermenge"}, Unit: "%"},
		BoilerSetTemperature:       StatusField[int]{Id: "BoilerSetTemperature", Name: MultiLanguageString{EN: "Boiler Set Temperature", DE: "Kesselsolltemperatur"}, Unit: "°C"},
		CurrentUnderpressure:       StatusField[float64]{Id: "CurrentUnderpressure", Name: MultiLanguageString{EN: "Current Underpressure", DE: "Aktueller Unterdruck"}, Unit: "Pa"},
		AverageUnderpressure:       StatusField[float64]{Id: "AverageUnderpressure", Name: MultiLanguageString{EN: "Average Underpressure", DE: "Durchschnittlicher Unterdruck"}, Unit: "Pa"},
		SetUnderpressure:           StatusField[float64]{Id: "SetUnderpressure", Name: MultiLanguageString{EN: "Set Underpressure", DE: "Soll-Unterdruck"}, Unit: "Pa"},
		FlowTemperatureCircuit3:    StatusField[float64]{Id: "FlowTemperatureCircuit3", Name: MultiLanguageString{EN: "Flow Temperature Circuit 3", DE: "Vorlauftemperatur Kreis 3"}, Unit: "°C"},
		FlowTemperatureCircuit4:    StatusField[float64]{Id: "FlowTemperatureCircuit4", Name: MultiLanguageString{EN: "Flow Temperature Circuit 4", DE: "Vorlauftemperatur Kreis 4"}, Unit: "°C"},
		FlowTemperatureCircuit3Set: StatusField[float64]{Id: "FlowTemperatureCircuit3Set", Name: MultiLanguageString{EN: "Flow Temperature Circuit 3 Set", DE: "Soll-Vorlauftemperatur Kreis 3"}, Unit: "°C"},
		FlowTemperatureCircuit4Set: StatusField[float64]{Id: "FlowTemperatureCircuit4Set", Name: MultiLanguageString{EN: "Flow Temperature Circuit 4 Set", DE: "Soll-Vorlauftemperatur Kreis 4"}, Unit: "°C"},
		BoilerTemperature2SM:       StatusField[float64]{Id: "BoilerTemperature2SM", Name: MultiLanguageString{EN: "Boiler Temperature 2 SM", DE: "Kesseltemperatur 2 SM"}, Unit: "°C"},
		HK1FR25:                    StatusField[float64]{Id: "HK1FR25", Name: MultiLanguageString{EN: "HK1 FR25", DE: "HK1 FR25"}, Unit: "°C"},
		HK2FR25:                    StatusField[float64]{Id: "HK2FR25", Name: MultiLanguageString{EN: "HK2 FR25", DE: "HK2 FR25"}, Unit: "°C"},
		HK3FR25SM:                  StatusField[float64]{Id: "HK3FR25SM", Name: MultiLanguageString{EN: "HK3 FR25 SM", DE: "HK3 FR25 SM"}, Unit: "°C"},
		HK4FR25SM:                  StatusField[float64]{Id: "HK4FR25SM", Name: MultiLanguageString{EN: "HK4 FR25 SM", DE: "HK4 FR25 SM"}, Unit: "°C"},
		MotorCurrentFeedScrew:      StatusField[float64]{Id: "MotorCurrentFeedScrew", Name: MultiLanguageString{EN: "Motor Current Feed Screw", DE: "Motorstrom Förderschnecke"}, Unit: "A"},
		MotorCurrentAshDischarge:   StatusField[float64]{Id: "MotorCurrentAshDischarge", Name: MultiLanguageString{EN: "Motor Current Ash Discharge", DE: "Motorstrom Ascheaustragung"}, Unit: "A"},
		MotorCurrentRoomDischarge:  StatusField[float64]{Id: "MotorCurrentRoomDischarge", Name: MultiLanguageString{EN: "Motor Current Room Discharge", DE: "Motorstrom Raumaustragung"}, Unit: "A"},
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
	switch any(field.Value).(type) {
	case int:
		parsedValue, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("invalid int field[%d] [%s]: %v (fields: %s)", index, field.Id, err, strings.Join(fields, "|"))
			return
		}
		field.Value = any(parsedValue).(T)
	case float64:
		parsedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Printf("invalid float field[%d] [%s]: %v (fields: %s)", index, field.Id, err, strings.Join(fields, "|"))
			return
		}
		field.Value = any(parsedValue).(T)
	case string:
		field.Value = any(value).(T)
	default:
		log.Fatalf("unsupported consumer type for field %s at %d", field.Id, index)
	}
}

func parseStatusRecord(fields []string) (*StatusRecord, error) {
	if len(fields) < 32 {
		return nil, fmt.Errorf("not enough fields")
	}

	record := newEmptyStatusRecord()

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
	parseField(fields, 19, record.FlowTemperatureCircuit3)
	parseField(fields, 20, record.FlowTemperatureCircuit4)
	parseField(fields, 21, record.FlowTemperatureCircuit3Set)
	parseField(fields, 22, record.FlowTemperatureCircuit4Set)
	parseField(fields, 23, record.BoilerTemperature2SM)
	parseField(fields, 24, record.HK1FR25)
	parseField(fields, 25, record.HK2FR25)
	parseField(fields, 26, record.HK3FR25SM)
	parseField(fields, 27, record.HK4FR25SM)
	// 28, 29, 30, 31 are not used
	parseField(fields, 29, record.MotorCurrentFeedScrew)
	parseField(fields, 30, record.MotorCurrentAshDischarge)
	parseField(fields, 31, record.MotorCurrentRoomDischarge)

	return record, nil
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

func publishStatusRecord(device *homie.Device, record *StatusRecord) {
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_RAUCHGAS_TEMPERATUR).Set(record.ExhaustGasTemperature)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_BOILER1_TEMPERATUR).Set(record.BoilerTemperature1)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_AUSSEN_TEMPERATUR_AKTUELL).Set(record.CurrentOutdoorTemperature)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_AUSSEN_TEMPERATUR_GEMITTELT).Set(record.AverageOutdoorTemperature)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_KESSEL_TEMPERATUR).Set(record.BoilerTemperature)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_KESSEL_SOLL_TEMPERATUR).Set(record.BoilerSetTemperature)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_SAUGLUFT_GEBLÄSE).Set(record.ExhaustFan)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_PRIMÄR_LUFT_GEBLÄSE).Set(record.PrimaryAirFan)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_O2_IN_ABGAS).Set(record.O2InExhaustGas)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_FÖRDER_MENGE).Set(record.FeedRate)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_STROM_RAUMAUSTRAGUNG).Set(record.MotorCurrentRoomDischarge)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_STROM_ASCHEAUSTRAGUNG).Set(record.MotorCurrentAshDischarge)
	device.Node(NODE_PROCESSWERTE).Property(PROPERTY_STROM_EINSCHUB).Set(record.MotorCurrentFeedScrew)

	device.Node(NODE_HEIZKREIS1).Property(PROPERTY_VORLAUF_TEMPERATUR).Set(record.FlowTemperatureCircuit1)
	device.Node(NODE_HEIZKREIS1).Property(PROPERTY_VORLAUF_SOLL_TEMPERATUR).Set(record.FlowTemperatureCircuit1Set)
	device.Node(NODE_HEIZKREIS2).Property(PROPERTY_VORLAUF_TEMPERATUR).Set(record.FlowTemperatureCircuit2)
	device.Node(NODE_HEIZKREIS2).Property(PROPERTY_VORLAUF_SOLL_TEMPERATUR).Set(record.FlowTemperatureCircuit2Set)
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

	homieDevice = homie.
		NewDevice(DEVICE_NAME, "Hargassner Heizung").
		AddNode(NODE_PROCESSWERTE, "Prozesswerte", "Prozesswerte").
		AddProperty(PROPERTY_RAUCHGAS_TEMPERATUR, "Rauchgas Temperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_BOILER1_TEMPERATUR, "Boiler 1 Temperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_AUSSEN_TEMPERATUR_AKTUELL, "Aussentemperatur aktuell", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_AUSSEN_TEMPERATUR_GEMITTELT, "Aussentemperatur gemittelt", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_KESSEL_TEMPERATUR, "Kesseltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_KESSEL_SOLL_TEMPERATUR, "Kesselsolltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_MELDUNG, "Meldung", homie.TypeString).Node().
		AddProperty(PROPERTY_SAUGLUFT_GEBLÄSE, "Saugluftgebläse", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_PRIMÄR_LUFT_GEBLÄSE, "Primärluftgebläse", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_O2_IN_ABGAS, "O2 in Abgas", homie.TypeFloat).SetUnit("%").Node().
		AddProperty(PROPERTY_FÖRDER_MENGE, "Fördermenge", homie.TypeFloat).SetUnit("%").Node().
		AddProperty(PROPERTY_STROM_RAUMAUSTRAGUNG, "Strom Raumaustragung", homie.TypeFloat).SetUnit("A").Node().
		AddProperty(PROPERTY_STROM_ASCHEAUSTRAGUNG, "Strom Ascheaustragung", homie.TypeFloat).SetUnit("A").Node().
		AddProperty(PROPERTY_STROM_EINSCHUB, "Strom Einschub", homie.TypeFloat).SetUnit("A").Node().
		AddProperty(PROPERTY_STROM_EINSCHUB, "Strom Einschub", homie.TypeFloat).SetUnit("A").Node().
		AddProperty(PROPERTY_MELDUNG, "Meldung", homie.TypeString).Node().
		Device().
		AddNode("heizkreis1", "Heizkreis 1", "Heizkreis 1").
		AddProperty(PROPERTY_VORLAUF_TEMPERATUR, "Vorlauftemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_VORLAUF_SOLL_TEMPERATUR, "Vorlauf Solltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		Device().
		AddNode("heizkreis2", "Heizkreis 2", "Heizkreis 2").
		AddProperty(PROPERTY_VORLAUF_TEMPERATUR, "Vorlauftemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty(PROPERTY_VORLAUF_SOLL_TEMPERATUR, "Vorlauf Solltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		Device().
		AddNode(NODE_STOERUNG, "Störung", "Störung").
		AddProperty(PROPERTY_NR, "Nummer", homie.TypeInteger).Node().
		AddProperty(PROPERTY_TEXT, "Text", homie.TypeString).Node().
		AddProperty(PROPERTY_ACTIVE, "Aktiv", homie.TypeBoolean).Node().
		AddProperty(PROPERTY_LAST_CHANGE, "Letzte Änderung", homie.TypeString).Node().
		Device()

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
			record, err := parseStatusRecord(fields)
			if err != nil {
				log.Println("Error parsing status record:", err)
				continue
			}
			publishStatusRecord(homieDevice, record)
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

		homieDevice.Node(NODE_STOERUNG).Property(PROPERTY_NR).Set(stoerNr)
		homieDevice.Node(NODE_STOERUNG).Property(PROPERTY_TEXT).Set(stoerungText)
		homieDevice.Node(NODE_STOERUNG).Property(PROPERTY_ACTIVE).Set(active)
		homieDevice.Node(NODE_STOERUNG).Property(PROPERTY_LAST_CHANGE).Set(lastChange)

	} else {
		message := strings.Join(fields[2:], " ")
		homieDevice.Node(NODE_PROCESSWERTE).Property(PROPERTY_MELDUNG).Set(message)
	}
}
