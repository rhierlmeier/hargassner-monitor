package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.bug.st/serial"
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
		} else {
			fmt.Print(line)
		}
	}

}
