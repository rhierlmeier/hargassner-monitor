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

var homieDevice *homie.Device
var statusNode *homie.Node
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

func publishStatusRecord(record *StatusRecord) {
	statusNode.Property("PrimaryAirFan").Set(fmt.Sprintf("%d", record.PrimaryAirFan))
	statusNode.Property("ExhaustFan").Set(fmt.Sprintf("%d", record.ExhaustFan))
	statusNode.Property("O2InExhaustGas").Set(fmt.Sprintf("%f", record.O2InExhaustGas))
	statusNode.Property("BoilerTemperature").Set(fmt.Sprintf("%d", record.BoilerTemperature))
	statusNode.Property("ExhaustGasTemperature").Set(fmt.Sprintf("%d", record.ExhaustGasTemperature))
	statusNode.Property("CurrentOutdoorTemperature").Set(fmt.Sprintf("%f", record.CurrentOutdoorTemperature))
	statusNode.Property("AverageOutdoorTemperature").Set(fmt.Sprintf("%f", record.AverageOutdoorTemperature))
	statusNode.Property("FlowTemperatureCircuit1").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit1))
	statusNode.Property("FlowTemperatureCircuit2").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit2))
	statusNode.Property("FlowTemperatureCircuit1Set").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit1Set))
	statusNode.Property("FlowTemperatureCircuit2Set").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit2Set))
	statusNode.Property("ReturnBoiler2BufferTemp").Set(fmt.Sprintf("%d", record.ReturnBoiler2BufferTemp))
	statusNode.Property("BoilerTemperature1").Set(fmt.Sprintf("%d", record.BoilerTemperature1))
	statusNode.Property("FeedRate").Set(fmt.Sprintf("%d", record.FeedRate))
	statusNode.Property("BoilerSetTemperature").Set(fmt.Sprintf("%d", record.BoilerSetTemperature))
	statusNode.Property("CurrentUnderpressure").Set(fmt.Sprintf("%f", record.CurrentUnderpressure))
	statusNode.Property("AverageUnderpressure").Set(fmt.Sprintf("%f", record.AverageUnderpressure))
	statusNode.Property("SetUnderpressure").Set(fmt.Sprintf("%f", record.SetUnderpressure))
	statusNode.Property("FlowTemperatureCircuit3").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit3))
	statusNode.Property("FlowTemperatureCircuit4").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit4))
	statusNode.Property("FlowTemperatureCircuit3Set").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit3Set))
	statusNode.Property("FlowTemperatureCircuit4Set").Set(fmt.Sprintf("%f", record.FlowTemperatureCircuit4Set))
	statusNode.Property("BoilerTemperature2SM").Set(fmt.Sprintf("%f", record.BoilerTemperature2SM))
	statusNode.Property("HK1FR25").Set(fmt.Sprintf("%f", record.HK1FR25))
	statusNode.Property("HK2FR25").Set(fmt.Sprintf("%f", record.HK2FR25))
	statusNode.Property("HK3FR25SM").Set(fmt.Sprintf("%f", record.HK3FR25SM))
	statusNode.Property("HK4FR25SM").Set(fmt.Sprintf("%f", record.HK4FR25SM))
	statusNode.Property("BoilerState").Set(fmt.Sprintf("%f", record.BoilerState))
	statusNode.Property("MotorCurrentFeedScrew").Set(fmt.Sprintf("%f", record.MotorCurrentFeedScrew))
	statusNode.Property("MotorCurrentAshDischarge").Set(fmt.Sprintf("%f", record.MotorCurrentAshDischarge))
	statusNode.Property("MotorCurrentRoomDischarge").Set(fmt.Sprintf("%f", record.MotorCurrentRoomDischarge))
}

func onSet(topic, value string, dataType homie.PropertyType) {
	if value == "<nil>" {
		value = ""
	}
	if value == "" && dataType != homie.TypeString {
		// don't send a blank string on anything else than a string data type
		return
	}
	mqttClient.Publish(topic, 0, false, value)
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

	homieDevice = homie.NewDevice("hargassner-monitor", "Hargassner Monitor")
	statusNode = homieDevice.AddNode("status", "Status", "status")

	statusNode.AddProperty("PrimaryAirFan", "Primary Air Fan", "integer")
	statusNode.AddProperty("ExhaustFan", "Exhaust Fan", "integer")
	statusNode.AddProperty("O2InExhaustGas", "O2 In Exhaust Gas", "float")
	statusNode.AddProperty("BoilerTemperature", "Boiler Temperature", "integer")
	statusNode.AddProperty("ExhaustGasTemperature", "Exhaust Gas Temperature", "integer")
	statusNode.AddProperty("CurrentOutdoorTemperature", "Current Outdoor Temperature", "float")
	statusNode.AddProperty("AverageOutdoorTemperature", "Average Outdoor Temperature", "float")
	statusNode.AddProperty("FlowTemperatureCircuit1", "Flow Temperature Circuit 1", "float")
	statusNode.AddProperty("FlowTemperatureCircuit2", "Flow Temperature Circuit 2", "float")
	statusNode.AddProperty("FlowTemperatureCircuit1Set", "Flow Temperature Circuit 1 Set", "float")
	statusNode.AddProperty("FlowTemperatureCircuit2Set", "Flow Temperature Circuit 2 Set", "float")
	statusNode.AddProperty("ReturnBoiler2BufferTemp", "Return Boiler 2 Buffer Temp", "integer")
	statusNode.AddProperty("BoilerTemperature1", "Boiler Temperature 1", "integer")
	statusNode.AddProperty("FeedRate", "Feed Rate", "integer")
	statusNode.AddProperty("BoilerSetTemperature", "Boiler Set Temperature", "integer")
	statusNode.AddProperty("CurrentUnderpressure", "Current Underpressure", "float")
	statusNode.AddProperty("AverageUnderpressure", "Average Underpressure", "float")
	statusNode.AddProperty("SetUnderpressure", "Set Underpressure", "float")
	statusNode.AddProperty("FlowTemperatureCircuit3", "Flow Temperature Circuit 3", "float")
	statusNode.AddProperty("FlowTemperatureCircuit4", "Flow Temperature Circuit 4", "float")
	statusNode.AddProperty("FlowTemperatureCircuit3Set", "Flow Temperature Circuit 3 Set", "float")
	statusNode.AddProperty("FlowTemperatureCircuit4Set", "Flow Temperature Circuit 4 Set", "float")
	statusNode.AddProperty("BoilerTemperature2SM", "Boiler Temperature 2 SM", "float")
	statusNode.AddProperty("HK1FR25", "HK1 FR25", "float")
	statusNode.AddProperty("HK2FR25", "HK2 FR25", "float")
	statusNode.AddProperty("HK3FR25SM", "HK3 FR25 SM", "float")
	statusNode.AddProperty("HK4FR25SM", "HK4 FR25 SM", "float")
	statusNode.AddProperty("BoilerState", "Boiler State", "float")
	statusNode.AddProperty("MotorCurrentFeedScrew", "Motor Current Feed Screw", "float")
	statusNode.AddProperty("MotorCurrentAshDischarge", "Motor Current Ash Discharge", "float")
	statusNode.AddProperty("MotorCurrentRoomDischarge", "Motor Current Room Discharge", "float")

	homieDevice.OnSet(onSet)

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
			publishStatusRecord(record)
		} else {
			fmt.Print(line)
		}
	}

}
