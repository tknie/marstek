package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/tknie/log"
	"github.com/tknie/marstek"
)

func init() {
	log.InitZapLogWithFilename("marstek.log")
}

func main() {
	list := false
	em := false
	es := false
	bat := false
	pv := false
	mode := false
	overview := false
	flag.BoolVar(&list, "d", false, "List all available devices")
	flag.BoolVar(&em, "e", false, "Show energy meter status")
	flag.BoolVar(&es, "s", false, "Show energy system status")
	flag.BoolVar(&bat, "b", false, "Show battery status")
	flag.BoolVar(&pv, "p", false, "Show photovoltaic status")
	flag.BoolVar(&mode, "m", false, "Show device mode")
	flag.BoolVar(&overview, "o", false, "Show device overview (mode, energy system, energy meter, battery and photovoltaic status)")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: client [options] <device>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	fmt.Println("Starting connecting to ... " + args[0])

	m := marstek.New(args[0])
	switch {
	case list:
		d, err := m.GetDevice("0")
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Device:")
		dumpMap(2, d)

	case bat:
		d, err := m.GetBatStatus()
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Battery:")
		dumpMap(2, d)
	case em:
		d, err := m.GetEnergyMeterStatus()
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Energy Meter:")
		dumpMap(2, d)
	case pv:
		d, err := m.GetPVStatus()
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Photovoltaic:")
		dumpMap(2, d)
	case es:
		d, err := m.GetEnergySystemStatus()
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Energy System:")
		dumpMap(2, d)
	case mode:
		d, err := m.GetMode()
		if err != nil {
			fmt.Println("Error getting device: " + err.Error())
			return
		}
		fmt.Println("Device Mode:")
		dumpMap(2, d)
	case overview:
		d, err := m.Summary()
		if err != nil {
			fmt.Println("Error getting device summary: " + err.Error())
			return
		}
		fmt.Println("Device Overview:")
		dumpMap(2, d)
	default:
		fmt.Println("No valid option provided. Use -l, -e, -s, -b or -p.")
	}
}

func dumpMap(level int, m map[string]interface{}) {
	prefix := strings.Repeat(" ", level)
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			fmt.Println(prefix + k + ":")
			dumpMap(level+2, val)
		default:
			fmt.Printf("%s%s = %v\n", prefix, k, v)
		}
	}
}
