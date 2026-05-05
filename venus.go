package marstek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"text/template"
	"time"

	"github.com/tknie/log"
)

var requestBaseJson = `{"id": {{.Id}},"method": "{{.Methods}}","params": {{.Params}}}`
var requestManualScheduleJson = `{"id": {{.Id}},"config": {"mode": "Manual","manual_cfg": {"time_num": {{.TimeNum}},"start_time": "{{.StartTime}}","end_time": "{{.EndTime}}","week_set": {{.WeekSet}},"power": {{.Power}},"enable": {{.Enable}}}}}`
var requestPassiveJson = `{"id": {{.Id}}, "config": {"mode": "Passive","passive_cfg": { "power": {{.Request}}, "cd_time": {{.SetTime}}}}}`

// MaxManualSchedules defines the maximum number of manual schedules that can be set for the device. This is based on the device's capabilities and ensures that we do not exceed the allowed number of schedules when clearing them.
const MaxManualSchedules = 10

// MaxRetries defines the maximum number of retry attempts when a read timeout occurs while waiting for a response from
// the device. This allows for multiple attempts to get a response in case of temporary issues with the device or network.
var MaxRetries = 4

// ReadTimeout defines the duration in seconds to wait for a response from the device before considering it a timeout. This
// is used to set the read deadline when waiting for a response, allowing for a reasonable amount of time for the device to
// respond before retrying or giving up.
var ReadTimeout = 15

// RetryDelay defines the delay in seconds between retry attempts when a read timeout
// occurs while waiting for a response from the device. This allows for a brief pause
// before retrying the request, which can help in cases where the device may be temporarily
// unresponsive or experiencing network issues.
var RetryDelay = 1

type Marstek struct {
	Services   string
	Connection net.Conn
}

type Schedule struct {
	Id        int
	Power     int
	Enable    bool
	TimeNum   int
	StartTime string
	EndTime   string
	WeekSet   int
}

var requestId uint64 = 1

// New creates a new instance of the Marstek struct with the provided
// service address. It initializes the Services field with the given
// address and returns a pointer to the Marstek instance. This function
// is used to create a new Marstek object that can be used to interact with
// the device.
func New(service string) *Marstek {
	return &Marstek{
		Services: service,
	}
}

// connect creates a UDP connection to the Marstek device using
// the provided service address. It resolves the UDP address and
// establishes a connection, which is stored in the Marstek struct
// for future communication. The method returns any error encountered
// during the connection process.
func (m *Marstek) connect() error {
	addr, err := net.ResolveUDPAddr("udp", m.Services)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	log.Log.Infof("Established connection to %s", m.Services)
	m.Connection = conn
	return nil
}

// sendMessage sends a byte array message to the Marstek device and waits
// for a response. It establishes a connection, sends the message, and reads
// the response from the device. The method handles timeouts and retries if
// necessary, and returns the response as a byte array along with any error
// encountered during the process.
func (m *Marstek) sendMessage(sendData []byte) ([]byte, error) {
	err := m.connect()
	if err != nil {
		return nil, err
	}
	defer m.Connection.Close()

	log.Log.Debugf("Sending request: %s", string(sendData))
	m.Connection.Write(sendData)
	log.Log.Infof("Request sent, waiting for response...")
	b := make([]byte, 4096)
	m.Connection.SetReadDeadline(time.Now().Add(time.Duration(ReadTimeout) * time.Second))
	n, err := m.Connection.Read(b)
	if err != nil {
		return nil, err
	}
	log.Log.Debugf("Received %d bytes\n", n)
	return b[:n], err
}

// sendRequest constructs a JSON request using a template and sends it to the Marstek
// device. It takes in the method name and parameters as input, generates the request
// JSON, and sends it using the sendMessage method. The function handles retries in
// case of timeouts and returns the response as a map along with any error encountered
// during the process.
func (m *Marstek) sendRequest(methods, params string) (map[string]interface{}, error) {
	if params == "" {
		params = `{"id": 0}`
	}
	tmpl, err := template.New("request").Parse(requestBaseJson)
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
			if result, ok := v["result"]; ok {
				return result.(map[string]interface{}), nil
			}
			if marstekError, ok := v["error"]; ok {
				errorInfo := marstekError.(map[string]interface{})
				info, err := json.Marshal(&errorInfo)
				if err != nil {
					return nil, fmt.Errorf("Marstek error: %s", marstekError)
				}
				return nil, fmt.Errorf("Device error: %s", string(info))
			}

			return nil, fmt.Errorf("Device result info not received")
		case ok && netErr.Timeout():
			log.Log.Infof("Read timeout, retrying... (%d/%d)", i+1, MaxRetries)
			time.Sleep(time.Duration(RetryDelay) * time.Second)
			continue
		default:
			log.Log.Errorf("Error reading from connection: %v", err)
			fmt.Printf("Error reading from connection: %v\n", err)
			return nil, err
		}
	}
	return nil, fmt.Errorf("Max tries reached without a successful response")
}

// GetDevice retrieves the device information by sending a request to the device using its
// BLE MAC address. It constructs a JSON request with the provided BLE MAC address and sends
// it to the device to obtain the device information. The method returns the device information
// as a map and any error encountered during the process.
func (m *Marstek) GetDevice(bleMac string) (map[string]interface{}, error) {
	return m.sendRequest("Marstek.GetDevice", fmt.Sprintf(`{"ble_mac":"%s"}`, bleMac))
}

// GetWifiStatus retrieves the Wi-Fi status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the
// Wi-Fi status information. The method returns the Wi-Fi status as a map and any error
// encountered during the process.
func (m *Marstek) GetWifiStatus() (map[string]interface{}, error) {
	return m.sendRequest("Wifi.GetStatus", "")
}

// GetBatStatus retrieves the battery status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the
// battery status information. The method returns the battery status as a map and any
// error encountered during the process.
func (m *Marstek) GetBatStatus() (map[string]interface{}, error) {
	return m.sendRequest("Bat.GetStatus", "")
}

// GetBluetoothStatus retrieves the Bluetooth status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the
// Bluetooth status information. The method returns the Bluetooth status as a map and any
// error encountered during the process.
func (m *Marstek) GetBluetoothStatus() (map[string]interface{}, error) {
	return m.sendRequest("BLE.GetStatus", "")
}

// GetPVStatus retrieves the photovoltaic (PV) status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the PV
// status information. The method returns the PV status as a map and any error encountered
// during the process.
func (m *Marstek) GetPVStatus() (map[string]interface{}, error) {
	return m.sendRequest("PV.GetStatus", "")
}

// GetEnergySystemStatus retrieves the energy system status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the energy
// system status information. The method returns the energy system status as a map and any
// error encountered during the process.
func (m *Marstek) GetEnergySystemStatus() (map[string]interface{}, error) {
	return m.sendRequest("ES.GetStatus", "")
}

// GetEnergyMeterStatus retrieves the energy meter status of the device by sending a request to
// the device. It constructs a JSON request and sends it to the device to obtain the energy
// meter status information. The method returns the energy meter status as a map and any
// error encountered during the process.
func (m *Marstek) GetEnergyMeterStatus() (map[string]interface{}, error) {
	return m.sendRequest("EM.GetStatus", "")
}

// GetMode retrieves the current mode of the device by sending a request to
// the device. It constructs a JSON request using a template and sends it
// to the device to obtain the mode information. The method returns the
// mode information as a map and any error encountered during the process.
func (m *Marstek) GetMode() (map[string]interface{}, error) {
	return m.sendRequest("ES.GetMode", "")
}

// SetDOD sets the depth of discharge (DOD) value for the device by sending a request to
// the device. It constructs a JSON request using a template and sends it to the device
// to update the DOD settings. The method takes in the desired DOD value as a parameter
// and returns any error encountered during the process. The DOD value must be between 33
// and 88, and the method will return an error if the value is out of range.
func (m *Marstek) SetDOD(value int) error {
	if value < 33 || value > 88 {
		return fmt.Errorf("DOD value must be between 33 and 88")
	}
	_, err := m.sendRequest("DOD.SET", fmt.Sprintf(`{"value": %d}}`, value))
	return err
}

// BluetoothLock sets the Bluetooth lock status of the device by sending a request to the device.
// It constructs a JSON request using a template and sends it to the device to update the Bluetooth
// lock settings. The method takes in a boolean parameter to enable or disable the Bluetooth lock
// and returns any error encountered during the process. The enable parameter determines whether to
// enable (true) or disable (false) the Bluetooth lock, and the method sends the appropriate request
// to the device based on the value of this parameter.
func (m *Marstek) BluetoothLock(enable bool) error {
	enableValue := 1
	if enable {
		enableValue = 0
	}
	_, err := m.sendRequest("BLE.Lock", fmt.Sprintf(`{"enable": %d}`, enableValue))
	return err
}

// LEDControl sets the LED control status of the device by sending a request to the device. It constructs
// a JSON request using a template and sends it to the device to update the LED control settings. The method
// takes in a boolean parameter to enable or disable the LED control and returns any error encountered during
// the process. The enable parameter determines whether to enable (true) or disable (false) the LED control,
// and the method sends the appropriate request to the device based on the value of this parameter.
func (m *Marstek) LEDControl(enable bool) error {
	enableValue := 1
	if !enable {
		enableValue = 0
	}
	_, err := m.sendRequest("Led.Ctrl", fmt.Sprintf(`{"state": %d}`, enableValue))
	return err
}

// PassivePowerConsumption sets the environment power consumption by
// sending a request to the device. It constructs a JSON request using a
// template and sends it to the device to update the power consumption settings.
// The method takes in the desired power level and cooldown time as parameters and
// returns any error encountered during the process.
// The negative power value indicates the power consumption level into the battery,
// positive power value indicates the power output level from the battery, and zero
// means no power consumption or output.
func (m *Marstek) PassivePowerConsumption(power, cdTime int) error {
	device, err := m.GetDevice("0")
	if err != nil {
		return err
	}
	log.Log.Infof("Device info: %v", device["id"])
	tmpl, err := template.New("request").Parse(requestPassiveJson)
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

	log.Log.Infof("Sending request: %s", buffer.String())
	_, err = m.sendRequest("ES.SetMode", buffer.String())
	return err
}

// SetAutoMode sets the device to auto mode by sending a request to the device.
// It constructs a JSON request using a template and sends it to the device to
// update the mode settings. The method returns any error encountered during the
// process.
func (m *Marstek) SetAutoMode(enable bool) error {
	enableValue := 1
	if !enable {
		enableValue = 0
	}
	mode := `{"id": 1,"config": {"mode": "Auto","auto_cfg": {"enable": ` + strconv.Itoa(enableValue) + `}}}}`
	_, err := m.sendRequest("ES.SetMode", mode)
	return err
}

// SetAIMode sets the device to AI mode by sending a request to the device. It
// constructs a JSON request using a template and sends it to the device to update
// the mode settings. The method takes in a boolean parameter to enable or disable
// AI mode and returns any error encountered during the process.
func (m *Marstek) SetAIMode(enable bool) error {
	enableValue := 1
	if !enable {
		enableValue = 0
	}
	mode := `{"id": 1,"config": {"mode": "Auto","ai_cfg": {"enable": ` + strconv.Itoa(enableValue) + `}}}}`
	_, err := m.sendRequest("ES.SetMode", mode)
	return err
}

// SetUPSMode sets the device to UPS mode by sending a request to the device. It
// constructs a JSON request using a template and sends it to the device to update
// the mode settings. The method takes in a boolean parameter to enable or disable
// UPS mode and returns any error encountered during the process.
func (m *Marstek) SetUPSMode(enable bool) error {
	enableValue := 1
	if !enable {
		enableValue = 0
	}
	mode := `{"id": 1,"config": {"mode": "Auto","ups_cfg": {"enable": ` + strconv.Itoa(enableValue) + `}}}}`
	_, err := m.sendRequest("ES.SetMode", mode)
	return err
}

// SetManualSchedule sets a manual schedule for the device by sending a request to
// the device. It constructs a JSON request using a template and sends it to the device
// to update the manual schedule settings. The method takes in a Schedule struct as a
// parameter, which contains the schedule details, and returns any error encountered
// during the process.
func (m *Marstek) SetManualSchedule(schedule *Schedule) error {
	tmpl, err := template.New("schedule").Parse(requestManualScheduleJson)
	if err != nil {
		log.Log.Errorf("Error generating from template: %v", err)
		return err
	}
	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, schedule)
	if err != nil {
		return err
	}

	_, err = m.sendRequest("ES.SetMode", buffer.String())
	if err != nil {
		return err
	}
	return nil
}

// ClearManualSchedule clears all manual schedules by sending a request to set
// each schedule to disabled with default values. It iterates through the maximum
// number of allowed manual schedules and sends a request for each one to ensure
// they are all cleared.
func (m *Marstek) ClearManualSchedule() error {
	tmpl, err := template.New("schedule").Parse(requestManualScheduleJson)
	if err != nil {
		log.Log.Errorf("Error generating from template: %v", err)
		return err
	}
	schedule := &Schedule{1, 0, false, 0, "00:00", "00:00", 0}
	for i := 0; i < MaxManualSchedules; i++ {
		var buffer bytes.Buffer
		schedule.TimeNum = i
		err = tmpl.Execute(&buffer, schedule)
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

func (m *Marstek) Summary() (map[string]interface{}, error) {
	marstekMap := make(map[string]interface{})
	device, err := m.GetDevice("0")
	if err != nil {
		log.Log.Errorf("Error getting device parameter: %v", err)
		return nil, err
	}
	marstekMap["device"] = device
	battery, err := m.GetBatStatus()
	if err != nil {
		log.Log.Errorf("Error getting battery status: %v", err)
		return nil, err
	}
	marstekMap["battery"] = battery
	meterStatus, err := m.GetEnergyMeterStatus()
	if err != nil {
		log.Log.Errorf("Error getting energy meter status: %v", err)
		return nil, err
	}
	marstekMap["energy_meter"] = meterStatus
	systemStatus, err := m.GetEnergySystemStatus()
	if err != nil {
		log.Log.Errorf("Error getting energy system status: %v", err)
		return nil, err
	}
	marstekMap["energy_system"] = systemStatus
	pvStatus, err := m.GetPVStatus()
	if err != nil {
		log.Log.Errorf("Error photovoltaic system status: %v", err)
		return nil, err
	}
	marstekMap["photovoltaic"] = pvStatus
	return marstekMap, nil
}
