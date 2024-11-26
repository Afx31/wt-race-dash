package main

import (
	"encoding/json"
	"log"
	"time"
)

type WarningAlerts struct {
	Type             int8
	AlertCoolantTemp bool
	AlertOilTemp     bool
	AlertOilPressure bool
}

var (
	warningAlerts = WarningAlerts{Type: 4}
)

func (wsConn *MySocket) HandleWarningAlerts() {
	for {
		warningAlerts.AlertCoolantTemp = int(canData.Ect) > appSettings.WarningValues["warningCoolantTemp"]
		warningAlerts.AlertOilTemp = int(canData.OilTemp) > appSettings.WarningValues["warningOilTemp"]
		warningAlerts.AlertOilPressure = int(canData.OilPressure) > appSettings.WarningValues["warningOilPressure"]

		jsonData, err := json.Marshal(warningAlerts)
		if err != nil {
			log.Println("Json Marshal error (Warning Alerts): ", err)
			return
		}
    
		wsConn.writeToClient(warningAlerts.Type, jsonData)
		time.Sleep(5 * time.Second)
	}
}