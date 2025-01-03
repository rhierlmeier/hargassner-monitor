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

type StatusRecord struct {
	PrimaryAirFan              int
	ExhaustFan                 int
	O2InExhaustGas             float64
	BoilerTemperature          int
	ExhaustGasTemperature      int
	CurrentOutdoorTemperature  float64
	AverageOutdoorTemperature  float64
	FlowTemperatureCircuit1    float64
	FlowTemperatureCircuit2    float64
	FlowTemperatureCircuit1Set float64
	FlowTemperatureCircuit2Set float64
	ReturnBoiler2BufferTemp    int
	BoilerTemperature1         int
	FeedRate                   int
	BoilerSetTemperature       int
	CurrentUnderpressure       float64
	AverageUnderpressure       float64
	SetUnderpressure           float64
	FlowTemperatureCircuit3    float64
	FlowTemperatureCircuit4    float64
	FlowTemperatureCircuit3Set float64
	FlowTemperatureCircuit4Set float64
	BoilerTemperature2SM       float64
	HK1FR25                    float64
	HK2FR25                    float64
	HK3FR25SM                  float64
	HK4FR25SM                  float64
	BoilerState                float64
	MotorCurrentFeedScrew      float64
	MotorCurrentAshDischarge   float64
	MotorCurrentRoomDischarge  float64
}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is ready"))
}

func parseStatusRecord(fields []string) (*StatusRecord, error) {
	if len(fields) < 32 {
		return nil, fmt.Errorf("not enough fields")
	}

	record := &StatusRecord{}
	var err error

	record.PrimaryAirFan, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "PrimaryAirFan", 1, err)
	}
	record.ExhaustFan, err = strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "ExhaustFan", 2, err)
	}
	record.O2InExhaustGas, err = strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "O2InExhaustGas", 3, err)
	}
	record.BoilerTemperature, err = strconv.Atoi(fields[4])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "BoilerTemperature", 4, err)
	}
	record.ExhaustGasTemperature, err = strconv.Atoi(fields[5])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "ExhaustGasTemperature", 5, err)
	}
	record.CurrentOutdoorTemperature, err = strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "CurrentOutdoorTemperature", 6, err)
	}
	record.AverageOutdoorTemperature, err = strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "AverageOutdoorTemperature", 7, err)
	}
	record.FlowTemperatureCircuit1, err = strconv.ParseFloat(fields[8], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit1", 8, err)
	}
	record.FlowTemperatureCircuit2, err = strconv.ParseFloat(fields[9], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit2", 9, err)
	}
	record.FlowTemperatureCircuit1Set, err = strconv.ParseFloat(fields[10], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit1Set", 10, err)
	}
	record.FlowTemperatureCircuit2Set, err = strconv.ParseFloat(fields[11], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit2Set", 11, err)
	}
	record.ReturnBoiler2BufferTemp, err = strconv.Atoi(fields[12])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "ReturnBoiler2BufferTemp", 12, err)
	}
	record.BoilerTemperature1, err = strconv.Atoi(fields[13])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "BoilerTemperature1", 13, err)
	}
	record.FeedRate, err = strconv.Atoi(fields[14])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FeedRate", 14, err)
	}
	record.BoilerSetTemperature, err = strconv.Atoi(fields[15])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "BoilerSetTemperature", 15, err)
	}
	record.CurrentUnderpressure, err = strconv.ParseFloat(fields[16], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "CurrentUnderpressure", 16, err)
	}
	record.AverageUnderpressure, err = strconv.ParseFloat(fields[17], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "AverageUnderpressure", 17, err)
	}
	record.SetUnderpressure, err = strconv.ParseFloat(fields[18], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "SetUnderpressure", 18, err)
	}
	record.FlowTemperatureCircuit3, err = strconv.ParseFloat(fields[19], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit3", 19, err)
	}
	record.FlowTemperatureCircuit4, err = strconv.ParseFloat(fields[20], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit4", 20, err)
	}
	record.FlowTemperatureCircuit3Set, err = strconv.ParseFloat(fields[21], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit3Set", 21, err)
	}
	record.FlowTemperatureCircuit4Set, err = strconv.ParseFloat(fields[22], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "FlowTemperatureCircuit4Set", 22, err)
	}
	record.BoilerTemperature2SM, err = strconv.ParseFloat(fields[23], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "BoilerTemperature2SM", 23, err)
	}
	record.HK1FR25, err = strconv.ParseFloat(fields[24], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "HK1FR25", 24, err)
	}
	record.HK2FR25, err = strconv.ParseFloat(fields[25], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "HK2FR25", 25, err)
	}
	record.HK3FR25SM, err = strconv.ParseFloat(fields[26], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "HK3FR25SM", 26, err)
	}
	record.HK4FR25SM, err = strconv.ParseFloat(fields[27], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "HK4FR25SM", 27, err)
	}
	record.BoilerState, err = strconv.ParseFloat(fields[28], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "BoilerState", 28, err)
	}
	record.MotorCurrentFeedScrew, err = strconv.ParseFloat(fields[29], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "MotorCurrentFeedScrew", 29, err)
	}
	record.MotorCurrentAshDischarge, err = strconv.ParseFloat(fields[30], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "MotorCurrentAshDischarge", 30, err)
	}
	record.MotorCurrentRoomDischarge, err = strconv.ParseFloat(fields[31], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %v", "MotorCurrentRoomDischarge", 31, err)
	}

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
		os.Exit(1)
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
		os.Exit(1)
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
		fields := strings.Fields(strings.TrimSpace(line))

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
			log.Fatalf("Unexpected value of fields[3] (%s): %s (line: %s)", fields[3], "Set or Quit", line)
			return
		}

		stoerNr, err := strconv.Atoi(fields[4])
		if err != nil {
			log.Printf("Unexpected value of fields[4] (%s): %s (line: %s)", fields[4], err, line)
			return
		}
		stoerungText := getStoerungText(stoerNr)

		if active {
			log.Fatalf("Störung %d: %s", stoerNr, stoerungText)
		} else {
			stoerungText = ""
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
