package marstek

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4/testutils/assert"
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
	fmt.Printf("Response: %v\n", x)
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
	err = m.SetEnvironmentPowerConsumption(0, 3600)
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
