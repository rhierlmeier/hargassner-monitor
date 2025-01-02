package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.bug.st/serial"
)

func printUsage() {
	fmt.Println("Usage: hargassner-monitor [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

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

func parseStatusRecord(fields []string) (*StatusRecord, error) {

	if len(fields) < 31 {
		return nil, fmt.Errorf("invalid status record: %s", fields)
	}

	record := &StatusRecord{}
	var err error

	record.PrimaryAirFan, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "PrimaryAirFan", 1, err)
	}
	record.ExhaustFan, err = strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "ExhaustFan", 2, err)
	}
	record.O2InExhaustGas, err = strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "O2InExhaustGas", 3, err)
	}
	record.BoilerTemperature, err = strconv.Atoi(fields[4])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "BoilerTemperature", 4, err)
	}
	record.ExhaustGasTemperature, err = strconv.Atoi(fields[5])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "ExhaustGasTemperature", 5, err)
	}
	record.CurrentOutdoorTemperature, err = strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "CurrentOutdoorTemperature", 6, err)
	}
	record.AverageOutdoorTemperature, err = strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "AverageOutdoorTemperature", 7, err)
	}
	record.FlowTemperatureCircuit1, err = strconv.ParseFloat(fields[8], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit1", 8, err)
	}
	record.FlowTemperatureCircuit2, err = strconv.ParseFloat(fields[9], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit2", 9, err)
	}
	record.FlowTemperatureCircuit1Set, err = strconv.ParseFloat(fields[10], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit1Set", 10, err)
	}
	record.FlowTemperatureCircuit2Set, err = strconv.ParseFloat(fields[11], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit2Set", 11, err)
	}
	record.ReturnBoiler2BufferTemp, err = strconv.Atoi(fields[12])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "ReturnBoiler2BufferTemp", 12, err)
	}
	record.BoilerTemperature1, err = strconv.Atoi(fields[13])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "BoilerTemperature1", 13, err)
	}
	record.FeedRate, err = strconv.Atoi(fields[14])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FeedRate", 14, err)
	}
	record.BoilerSetTemperature, err = strconv.Atoi(fields[15])
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "BoilerSetTemperature", 15, err)
	}
	record.CurrentUnderpressure, err = strconv.ParseFloat(fields[16], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "CurrentUnderpressure", 16, err)
	}
	record.AverageUnderpressure, err = strconv.ParseFloat(fields[17], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "AverageUnderpressure", 17, err)
	}
	record.SetUnderpressure, err = strconv.ParseFloat(fields[18], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "SetUnderpressure", 18, err)
	}
	record.FlowTemperatureCircuit3, err = strconv.ParseFloat(fields[19], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit3", 19, err)
	}
	record.FlowTemperatureCircuit4, err = strconv.ParseFloat(fields[20], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit4", 20, err)
	}
	record.FlowTemperatureCircuit3Set, err = strconv.ParseFloat(fields[21], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit3Set", 21, err)
	}
	record.FlowTemperatureCircuit4Set, err = strconv.ParseFloat(fields[22], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "FlowTemperatureCircuit4Set", 22, err)
	}
	record.BoilerTemperature2SM, err = strconv.ParseFloat(fields[23], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "BoilerTemperature2SM", 23, err)
	}
	record.HK1FR25, err = strconv.ParseFloat(fields[24], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "HK1FR25", 24, err)
	}
	record.HK2FR25, err = strconv.ParseFloat(fields[25], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "HK2FR25", 25, err)
	}
	record.HK3FR25SM, err = strconv.ParseFloat(fields[26], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "HK3FR25SM", 26, err)
	}
	record.HK4FR25SM, err = strconv.ParseFloat(fields[27], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "HK4FR25SM", 27, err)
	}
	record.BoilerState, err = strconv.ParseFloat(fields[28], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "BoilerState", 28, err)
	}
	record.MotorCurrentFeedScrew, err = strconv.ParseFloat(fields[29], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "MotorCurrentFeedScrew", 29, err)
	}
	record.MotorCurrentAshDischarge, err = strconv.ParseFloat(fields[30], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "MotorCurrentAshDischarge", 30, err)
	}
	record.MotorCurrentRoomDischarge, err = strconv.ParseFloat(fields[31], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid field %s at %d: %g", "MotorCurrentRoomDischarge", 31, err)
	}

	return record, nil
}

func main() {
	portName := flag.String("port", "/dev/ttyAMA0", "Serial port name")
	baudRate := flag.Int("baud", 19200, "Baud rate for serial communication")

	dataBits := flag.Int("databits", 8, "Number of data bits for serial communication")
	parity := flag.String("parity", "none", "Parity for serial communication (none, even, odd)")
	var parityValue serial.Parity
	switch strings.ToUpper(*parity) {
	case "EVEN":
		parityValue = serial.EvenParity
	case "ODD":
		parityValue = serial.OddParity
	case "NONE":
		parityValue = serial.NoParity
	default:
		log.Fatal("Invalid parity: " + *parity)
		os.Exit(1)
	}
	stopBits := flag.String("stopbits", "1", "Number of stop bits for serial communication (1, 2, 1.5)")
	var stopBitValue serial.StopBits
	switch strings.ToUpper(*stopBits) {
	case "1":
		stopBitValue = serial.OneStopBit
	case "2":
		stopBitValue = serial.TwoStopBits
	case "1.5":
		stopBitValue = serial.OnePointFiveStopBits
	default:
		log.Fatal("Invalid stopbits: " + *stopBits)
		os.Exit(1)
	}

	flag.Usage = printUsage
	flag.Parse()

	mode := &serial.Mode{
		BaudRate: *baudRate,
		Parity:   parityValue,
		DataBits: *dataBits,
		StopBits: stopBitValue,
	}

	port, err := serial.Open(*portName, mode)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
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
