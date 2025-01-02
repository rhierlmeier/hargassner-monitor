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

func publishStatusRecord(device *homie.Device, record *StatusRecord) {
	device.Node("prozesswerte").Property("rauchgasTemperatur").Set(record.ExhaustGasTemperature)
	device.Node("prozesswerte").Property("boiler1Temperatur").Set(record.BoilerTemperature1)
	device.Node("prozesswerte").Property("aussenTemperaturAktuell").Set(record.CurrentOutdoorTemperature)
	device.Node("prozesswerte").Property("aussenTemperaturGemittelt").Set(record.AverageOutdoorTemperature)
	device.Node("prozesswerte").Property("kesselTemperatur").Set(record.BoilerTemperature)
	device.Node("prozesswerte").Property("kesselSollTemperatur").Set(record.BoilerSetTemperature)
	device.Node("prozesswerte").Property("saugluftGeblaese").Set(record.ExhaustFan)
	device.Node("prozesswerte").Property("primaerLuftGeblaese").Set(record.PrimaryAirFan)
	device.Node("prozesswerte").Property("o2InAbgas").Set(record.O2InExhaustGas)
	device.Node("prozesswerte").Property("foerderMenge").Set(record.FeedRate)
	device.Node("prozesswerte").Property("stromRaumaustragung").Set(record.MotorCurrentRoomDischarge)
	device.Node("prozesswerte").Property("stromAscheaustragung").Set(record.MotorCurrentAshDischarge)
	device.Node("prozesswerte").Property("stromEinschub").Set(record.MotorCurrentFeedScrew)
	device.Node("heizkreis1").Property("vorlaufTemperatur").Set(record.FlowTemperatureCircuit1)
	device.Node("heizkreis1").Property("vorlaufSollTemperatur").Set(record.FlowTemperatureCircuit1Set)
	device.Node("heizkreis2").Property("vorlaufTemperatur").Set(record.FlowTemperatureCircuit2)
	device.Node("heizkreis2").Property("vorlaufSollTemperatur").Set(record.FlowTemperatureCircuit2Set)
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
		log.Fatal(err)
		os.Exit(1)
	}

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

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	device := homie.
		NewDevice("hargassner", "Hargassner Heizung").
		AddNode("prozesswerte", "Prozesswerte", "Prozesswerte").
		AddProperty("rauchgasTemperatur", "Rauchgas Temperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("boiler1Temperatur", "Boiler 1 Temperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("aussenTemperaturAktuell", "Aussentemperatur aktuell", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("aussenTemperaturGemittelt", "Aussentemperatur gemittelt", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("kesselTemperatur", "Kesseltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("kesselSollTemperatur", "Kesselsolltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("meldung", "Meldung", homie.TypeString).Node().
		AddProperty("saugluftGeblaese", "Saugluftgebläse", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("primaerLuftGeblaese", "Primärluftgebläse", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("o2InAbgas", "O2 in Abgas", homie.TypeFloat).SetUnit("%").Node().
		AddProperty("foerderMenge", "Fördermenge", homie.TypeFloat).SetUnit("%").Node().
		AddProperty("stromRaumaustragung", "Strom Raumaustragung", homie.TypeFloat).SetUnit("A").Node().
		AddProperty("stromAscheaustragung", "Strom Ascheaustragung", homie.TypeFloat).SetUnit("A").Node().
		AddProperty("stromEinschub", "Strom Einschub", homie.TypeFloat).SetUnit("A").Node().
		AddProperty("stromEinschub", "Strom Einschub", homie.TypeFloat).SetUnit("A").Node().
		Device().
		AddNode("heizkreis1", "Heizkreis 1", "Heizkreis 1").
		AddProperty("vorlaufTemperatur", "Vorlauftemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("vorlaufSollTemperatur", "Vorlauf Solltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		Device().
		AddNode("heizkreis2", "Heizkreis 2", "Heizkreis 2").
		AddProperty("vorlaufTemperatur", "Vorlauftemperatur", homie.TypeFloat).SetUnit("°C").Node().
		AddProperty("vorlaufSollTemperatur", "Vorlauf Solltemperatur", homie.TypeFloat).SetUnit("°C").Node().
		Device().
		AddNode("stoerung", "Störung", "Störung").
		AddProperty("nr", "Nummer", homie.TypeInteger).Node().
		AddProperty("text", "Text", homie.TypeString).Node().Device()

	// get the full homie definition to send to MQTT - you only need to send it once unless it's changing over time
	for _, attribute := range device.GetHomieAttributes() {
		publish(attribute.Topic, attribute.Value)
	}

	device.OnSet(onSet)

	reader := bufio.NewReader(port)

	reader.ReadString('\n')

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fields := strings.Fields(strings.TrimSpace(line))

		if fields[0] == "pm" {
			record, err := parseStatusRecord(fields)
			if err != nil {
				log.Println("Error parsing status record:", err)
				continue
			}
			fmt.Printf("Parsed record: %+v\n", record)
			publishStatusRecord(device, record)

		} else {
			fmt.Print(line)
		}
	}

}
