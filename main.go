package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/tarm/serial"
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
	FlowTemperatureCircuit1Set int
	FlowTemperatureCircuit2Set int
	ReturnBoiler2BufferTemp    int
	BoilerTemperature1         int
	FeedRate                   int
	BoilerSetTemperature       int
	CurrentUnderpressure       float64
	AverageUnderpressure       float64
	SetUnderpressure           float64
	FlowTemperatureCircuit3    int
	FlowTemperatureCircuit4    int
	FlowTemperatureCircuit3Set int
	FlowTemperatureCircuit4Set int
	BoilerTemperature2SM       int
	HK1FR25                    int
	HK2FR25                    int
	HK3FR25SM                  int
	HK4FR25SM                  int
	BoilerState                int
	MotorCurrentFeedScrew      float64
	MotorCurrentAshDischarge   float64
	MotorCurrentRoomDischarge  float64
}

func parseStatusRecord(line string) (*StatusRecord, error) {
	fields := strings.Fields(line)
	if len(fields) < 31 {
		return nil, fmt.Errorf("invalid status record: %s", line)
	}

	record := &StatusRecord{}
	var err error

	record.PrimaryAirFan, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	record.ExhaustFan, err = strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}
	record.O2InExhaustGas, err = strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return nil, err
	}
	record.BoilerTemperature, err = strconv.Atoi(fields[4])
	if err != nil {
		return nil, err
	}
	record.ExhaustGasTemperature, err = strconv.Atoi(fields[5])
	if err != nil {
		return nil, err
	}
	record.CurrentOutdoorTemperature, err = strconv.ParseFloat(fields[6], 64)
	if err != nil {
		return nil, err
	}
	record.AverageOutdoorTemperature, err = strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit1, err = strconv.ParseFloat(fields[8], 64)
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit2, err = strconv.ParseFloat(fields[9], 64)
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit1Set, err = strconv.Atoi(fields[10])
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit2Set, err = strconv.Atoi(fields[11])
	if err != nil {
		return nil, err
	}
	record.ReturnBoiler2BufferTemp, err = strconv.Atoi(fields[12])
	if err != nil {
		return nil, err
	}
	record.BoilerTemperature1, err = strconv.Atoi(fields[13])
	if err != nil {
		return nil, err
	}
	record.FeedRate, err = strconv.Atoi(fields[14])
	if err != nil {
		return nil, err
	}
	record.BoilerSetTemperature, err = strconv.Atoi(fields[15])
	if err != nil {
		return nil, err
	}
	record.CurrentUnderpressure, err = strconv.ParseFloat(fields[16], 64)
	if err != nil {
		return nil, err
	}
	record.AverageUnderpressure, err = strconv.ParseFloat(fields[17], 64)
	if err != nil {
		return nil, err
	}
	record.SetUnderpressure, err = strconv.ParseFloat(fields[18], 64)
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit3, err = strconv.Atoi(fields[19])
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit4, err = strconv.Atoi(fields[20])
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit3Set, err = strconv.Atoi(fields[21])
	if err != nil {
		return nil, err
	}
	record.FlowTemperatureCircuit4Set, err = strconv.Atoi(fields[22])
	if err != nil {
		return nil, err
	}
	record.BoilerTemperature2SM, err = strconv.Atoi(fields[23])
	if err != nil {
		return nil, err
	}
	record.HK1FR25, err = strconv.Atoi(fields[24])
	if err != nil {
		return nil, err
	}
	record.HK2FR25, err = strconv.Atoi(fields[25])
	if err != nil {
		return nil, err
	}
	record.HK3FR25SM, err = strconv.Atoi(fields[26])
	if err != nil {
		return nil, err
	}
	record.HK4FR25SM, err = strconv.Atoi(fields[27])
	if err != nil {
		return nil, err
	}
	record.BoilerState, err = strconv.Atoi(fields[28])
	if err != nil {
		return nil, err
	}
	record.MotorCurrentFeedScrew, err = strconv.ParseFloat(fields[29], 64)
	if err != nil {
		return nil, err
	}
	record.MotorCurrentAshDischarge, err = strconv.ParseFloat(fields[30], 64)
	if err != nil {
		return nil, err
	}
	record.MotorCurrentRoomDischarge, err = strconv.ParseFloat(fields[31], 64)
	if err != nil {
		return nil, err
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
		parityValue = serial.ParityEven
	case "ODD":
		parityValue = serial.ParityOdd
	case "NONE":
		fallthrough
	default:
		parityValue = serial.ParityNone
	}
	stopBits := flag.Int("stopbits", 1, "Number of stop bits for serial communication")
	flag.Usage = printUsage
	flag.Parse()

	c := &serial.Config{
		Name:     *portName,
		Baud:     *baudRate,
		Size:     byte(*dataBits),
		Parity:   parityValue,
		StopBits: serial.StopBits(*stopBits),
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(s)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(line, "pm") {
			record, err := parseStatusRecord(line)
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
