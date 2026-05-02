package marstek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/tknie/log"
)

var REQUEST_JSON = `{"id": {{.Id}},"method": "{{.Methods}}","params": {{.Params}}}`
var REQUEST_MANUAL_SCHEDULE_JSON = `{"id": {{.Id}},"config": {"mode": "Manual","manual_cfg": {"time_num": {{.TimeNum}},"start_time": "{{.StartTime}}","end_time": "{{.EndTime}}","week_set": {{.WeekSet}},"power": {{.Power}},"enable": {{.Enable}}}}}`
var REQUEST_PASSIVE_JSON = `{"id": {{.Id}}, "config": {"mode": "Passive","passive_cfg": { "power": {{.Request}}, "cd_time": {{.SetTime}}}}}`

// MaxManualSchedules defines the maximum number of manual schedules that can be set for the device. This is based on the device's capabilities and ensures that we do not exceed the allowed number of schedules when clearing them.
const MaxManualSchedules = 10

type Marstek struct {
	Services   string
	Connection net.Conn
}

var requestId uint64 = 1

const MaxRetries = 10
const DefaultTimeout = 15

func New(service string) *Marstek {
	return &Marstek{
		Services: service,
	}
}

func (m *Marstek) connect() error {
	addr, err := net.ResolveUDPAddr("udp", m.Services)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	fmt.Printf("Established connection to %s \n", m.Services)
	m.Connection = conn
	return nil
}

func (m *Marstek) sendMessage(sendData []byte) ([]byte, error) {
	err := m.connect()
	if err != nil {
		return nil, err
	}
	defer m.Connection.Close()

	log.Log.Debugf("Sending request: %s", string(sendData))
	fmt.Println("Sending request...")
	m.Connection.Write(sendData)
	fmt.Println("Request sent, waiting for response...")
	b := make([]byte, 4096)
	m.Connection.SetReadDeadline(time.Now().Add(DefaultTimeout * time.Second))
	n, err := m.Connection.Read(b)
	if err != nil {
		return nil, err
	}
	log.Log.Debugf("Received %d bytes\n", n)
	return b[:n], err
}

func (m *Marstek) sendRequest(methods, params string) (map[string]interface{}, error) {
	if params == "" {
		params = `{"id": 0}`
	}
	tmpl, err := template.New("request").Parse(REQUEST_JSON)
	if err != nil {
		log.Log.Errorf("Error generating from template: %v", err)
		return nil, err
	}
	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, struct {
		Id      uint64
		Methods string
		Params  string
	}{requestId, methods, params})
	if err != nil {
		log.Log.Debugf("Error executing template: %v", err)
		return nil, err
	}
	for i := 0; i < MaxRetries; i++ {
		b, err := m.sendMessage(buffer.Bytes())

		netErr, ok := err.(net.Error)
		switch {
		case err == nil:
			v := make(map[string]interface{})
			err = json.Unmarshal(b, &v)
			if err != nil {
				log.Log.Errorf("Error unmarshaling JSON: %v\n", err)
				log.Log.Debugf("Received error data: %s\n", string(b))
				return nil, err
			}
			requestId++
			if log.IsDebugLevel() {
				buffer := bytes.Buffer{}
				json.Compact(&buffer, b)
				log.Log.Debugf("Received response: %s", buffer.String())
			}
			return v, nil
		case ok && netErr.Timeout():
			log.Log.Infof("Read timeout, retrying... (%d/%d)", i+1, MaxRetries)
			time.Sleep(1 * time.Second)
			continue
		default:
			log.Log.Errorf("Error reading from connection: %v", err)
			fmt.Printf("Error reading from connection: %v\n", err)
			return nil, err
		}
	}
	return nil, fmt.Errorf("Max tries reached without a successful response")
}

func (m *Marstek) GetDevice(bleMac string) (map[string]interface{}, error) {
	device, err := m.sendRequest("Marstek.GetDevice", fmt.Sprintf(`{"ble_mac":"%s"}`, bleMac))
	if err != nil {
		return nil, err
	}
	if result, ok := device["result"]; ok {
		return result.(map[string]interface{}), nil
	}
	if marstekError, ok := device["error"]; ok {
		errorInfo := marstekError.(map[string]interface{})
		info, err := json.Marshal(&errorInfo)
		if err != nil {
			return nil, fmt.Errorf("Device error: %v", marstekError)
		}
		return nil, fmt.Errorf("Device error: %v", info)
	}

	return nil, fmt.Errorf("Device result info not received")
}

func (m *Marstek) GetStatus() (map[string]interface{}, error) {
	return m.sendRequest("ES.GetStatus", "")
}

func (m *Marstek) GetMode() (map[string]interface{}, error) {
	return m.sendRequest("ES.GetMode", "")
}

func (m *Marstek) SetEnvironmentPowerConsumption(power, cdTime int) error {
	device, err := m.GetDevice("0")
	if err != nil {
		return err
	}
	fmt.Println("Device info:", device["id"])
	tmpl, err := template.New("request").Parse(REQUEST_PASSIVE_JSON)
	if err != nil {
		log.Log.Errorf("Error generating from template: %v", err)
		return err
	}
	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, struct {
		Id      uint64
		Request int
		SetTime int
	}{1, power, cdTime})
	if err != nil {
		panic(err)
	}

	fmt.Println("Sending request:", buffer.String())
	_, err = m.sendRequest("ES.SetMode", buffer.String())
	return err
}

func (m *Marstek) ClearManualSchedule() error {
	tmpl, err := template.New("schedule").Parse(REQUEST_MANUAL_SCHEDULE_JSON)
	if err != nil {
		log.Log.Errorf("Error generating from template: %v", err)
		return err
	}
	for i := 0; i < MaxManualSchedules; i++ {
		var buffer bytes.Buffer

		err = tmpl.Execute(&buffer, struct {
			Id        int
			Power     int
			Enable    bool
			TimeNum   int
			StartTime string
			EndTime   string
			WeekSet   int
		}{1, 0, false, i, "00:00", "00:00", 0})
		if err != nil {
			return err
		}

		_, err := m.sendRequest("ES.SetMode", buffer.String())
		if err != nil {
			return err
		}
	}
	return nil
}
