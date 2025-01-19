package main

import (
	"encoding/json"
	"log"

	"wt-race-dash/pkg/hondata"
)

type WarningAlerts struct {
	Type             int8
	AlertCoolantTemp bool
	AlertOilTemp     bool
	AlertOilPressure bool
}

var (
	warningAlerts = WarningAlerts{Type: 4}
	sendDataTrigger			bool
	previousCoolantTemp bool
	previousOilTemp     bool
	previousOilPressure bool
)

func (wsConn *MySocket) HandleWarningAlerts() {
	for {
		warningAlerts.AlertCoolantTemp = int(canInterface.(*hondata.CANFrameHandler).Frame661.Ect) > appSettings.WarningValues["warningCoolantTemp"]
		warningAlerts.AlertOilTemp = int(canInterface.(*hondata.CANFrameHandler).Frame667.OilTemp) > appSettings.WarningValues["warningOilTemp"]
		warningAlerts.AlertOilPressure = int(canInterface.(*hondata.CANFrameHandler).Frame667.OilPressure) > appSettings.WarningValues["warningOilPressure"]

		if (previousCoolantTemp != warningAlerts.AlertCoolantTemp) {
			sendDataTrigger = true
			previousCoolantTemp = !previousCoolantTemp
		}
		if (previousOilTemp != warningAlerts.AlertOilTemp) {
			sendDataTrigger = true
			previousOilTemp = !previousOilTemp
		}
		if (previousOilPressure != warningAlerts.AlertOilPressure) {
			sendDataTrigger = true
			previousOilPressure = !previousOilPressure
		}

		// Only send up IF it's a new value
		if (sendDataTrigger) {
			jsonData, err := json.Marshal(warningAlerts)
			if err != nil {
				log.Println("Json Marshal error (Warning Alerts): ", err)
				return
			}
			
			wsConn.writeToClient(warningAlerts.Type, jsonData)

			// Cleanup
			sendDataTrigger = false
		}
	}
}