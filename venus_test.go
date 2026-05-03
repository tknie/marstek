package marstek

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/testify/assert"
	"github.com/tknie/log"
	"go.uber.org/zap/zapcore"
)

func init() {
	log.InitZapLogLevelWithFilename("test.log", zapcore.DebugLevel)
}

func TestMastekDevice(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}

	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	x, err := m.sendRequest("Marstek.GetDevice",
		`{"ble_mac":"0"}`)
	if !assert.NoError(t, err, "Failed to send request Marstek.GetDevice") {
		return
	}
	x, err = m.sendRequest("Wifi.GetStatus", "")
	if !assert.NoError(t, err, "Failed to send request Wifi.GetStatus") {
		return
	}
	fmt.Printf("Response: %v\n", x)
}

func TestGetStatus(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}

	fmt.Println("Testing Marstek connection..." + t.Name())

	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	x, err := m.sendRequest("Wifi.GetStatus", "")
	if !assert.NoError(t, err, "Failed to send request Wifi.GetStatus") {
		return
	}
	x, err = m.sendRequest("Bat.GetStatus", "")
	if !assert.NoError(t, err, "Failed to send request Bat.GetStatus") {
		return
	}
	x, err = m.sendRequest("PV.GetStatus", "")
	if !assert.NoError(t, err, "Failed to send request PV.GetStatus") {
		return
	}
	fmt.Printf("Response: %v\n", x)
}

func TestEnergeStatus(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}

	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	x, err := m.sendRequest("Marstek.GetDevice",
		`{"ble_mac":"0"}`)
	if !assert.NoError(t, err, "Failed to send request Marstek.GetDevice") {
		return
	}
	fmt.Printf("Response: %v\n", x)
	x, err = m.sendRequest("ES.GetStatus", "")
	if !assert.NoError(t, err, "Failed to send request ES.GetStatus") {
		return
	}
	fmt.Printf("Response: %v\n", x)
	x, err = m.sendRequest("ES.GetMode", "")
	if !assert.NoError(t, err, "Failed to send request ES.GetMode") {
		return
	}
	fmt.Printf("Response: %v\n", x)
}

func TestGetEnergeMode(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}

	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	x, err := m.sendRequest("ES.GetMode", "")
	if !assert.NoError(t, err, "Failed to send request ES.GetMode") {
		return
	}

	fmt.Printf("Response:\n")
	dumpMap(2, x)
}

func dumpMap(level int, m map[string]interface{}) {
	prefix := strings.Repeat(" ", level)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			dumpMap(level+2, val)
		default:
			fmt.Printf("%s%s = %v\n", prefix, k, v)
		}
	}
}

func TestSetEnergeStatus(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}

	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	mode := `{
"id": 1,
"config": {
"mode": "Passive"
,
"passive_cfg": {
"power": -200,
"cd_time": 3000
}
}
}`
	x, err := m.sendRequest("ES.SetMode", mode)
	if !assert.NoError(t, err, "Failed to send request ES.SetMode") {
		return
	}

	fmt.Printf("Response: %v\n", x)
}

func TestSetMode(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}
	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	currentMode, err := m.GetMode()
	if !assert.NoError(t, err, "Failed to get mode") {
		return
	}
	info, err := json.Marshal(&currentMode)
	if !assert.NoError(t, err, "Failed to marshal mode data") {
		return
	}
	fmt.Printf("Current mode: %s\n", string(info))

	fmt.Printf("Connected to Marstek: %v\n", m)
	err = m.PassivePowerConsumption(0, 3600)
	if !assert.NoError(t, err, "Failed to set environment power consumption") {
		return
	}
	fmt.Printf("Set environment power consumption to Marstek, waiting ....\n")
	time.Sleep(1 * time.Second)
	data, err := m.GetMode()
	if !assert.NoError(t, err, "Failed to get mode") {
		return
	}
	info, err = json.Marshal(&data)
	if !assert.NoError(t, err, "Failed to marshal mode data") {
		return
	}
	fmt.Printf("Mode info: %s\n", string(info))
}

func TestClearSchedule(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}
	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	err := m.ClearManualSchedule()
	assert.NoError(t, err, "Failed to clear manual schedule")
}

func TestGetMode(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}
	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	result, err := m.GetEnergyMeterStatus()
	assert.NoError(t, err, "Failed to clear manual schedule")
	if assert.NotNil(t, result, "Result should not be nil") {
		fmt.Printf("Energy Meter Status:\n")
		dumpMap(0, result)
	}
}

func TestGetDevice(t *testing.T) {
	destination := os.Getenv("MARSTEK_DESTINATION")
	if destination == "" {
		fmt.Println("MARSTEK_DESTINATION environment variable is not set. Please set it to the Marstek device's address (e.g., 192.168.1.100:8080) and run the tests again.")
		return
	}
	fmt.Println("Testing Marstek connection..." + t.Name())
	m := New(destination)
	fmt.Printf("Connected to Marstek: %v\n", m)
	result, err := m.GetDevice("0")
	assert.NoError(t, err, "Failed to get device")
	if assert.NotNil(t, result, "Result should not be nil") {
		fmt.Printf("Device Info:\n")
		dumpMap(2, result)
	}
}
