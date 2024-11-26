package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os/exec"
	"sync"
	"time"

	"go.einride.tech/can/pkg/socketcan"
)

type CanData struct {
	Type             int8
	Rpm              uint16
	Speed            uint16
	Gear             uint8
	Voltage          float32
	Iat              uint16
	Ect              uint16
	Tps              uint16
	Map              uint16
	LambdaRatio      float64
	OilTemp          uint16
	OilPressure      uint16
  DataloggingAlert bool
}

var (
	canData       = CanData{Type: 1}
	
	// --- Data conversion constants ---
	// Oil Temp
	A = 0.0014222095
	B = 0.00023729017
	C = 9.3273998E-8
	// Oil Pressure
	originalLow  float64 = 0    //0.5
	originalHigh float64 = 5    //4.5
	desiredLow   float64 = -100 //0
	desiredHigh  float64 = 1100 //1000
)

func doDatalogging(dataloggingRunning *bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if *dataloggingRunning {
			fmt.Println("--- Datalogging: Finish ---")
			return
		}

		fmt.Println("--- Datalogging: Start ---")
		cmd := exec.Command("/home/pi/dev/wt-datalogging/bin/wt-datalogging")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error running datalogging: ", err)
		}

		fmt.Println(string(output))
	}
}

func (wsConn *MySocket) HandleCanBusData() {
	// ---------- CANBus data ----------
	canConn, _ := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
	defer canConn.Close()
	canRecv := socketcan.NewReceiver(canConn)
	isDatalogging := false

	for canRecv.Receive() {
		frame := canRecv.Frame()

		switch frame.ID {
		case 69, 105:
			wg.Add(1)
			go doDatalogging(&isDatalogging, &wg)
			time.Sleep(1 * time.Second)
			isDatalogging = !isDatalogging
			canData.DataloggingAlert = isDatalogging
		case 660, 1632:
			canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
			canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
			canData.Gear = frame.Data[4]
			canData.Voltage = float32(frame.Data[5]) / 10.0
		case 661, 1633:
			canData.Iat = binary.BigEndian.Uint16(frame.Data[0:2])
			canData.Ect = binary.BigEndian.Uint16(frame.Data[2:4])
		case 662, 1634:
			canData.Tps = binary.BigEndian.Uint16(frame.Data[0:2])
				if canData.Tps == 65535 { canData.Tps = 0	}
			canData.Map = binary.BigEndian.Uint16(frame.Data[2:4]) / 10
		case 664, 1636:
        canData.LambdaRatio = math.Round(float64(32768.0) / float64(binary.BigEndian.Uint16(frame.Data[0:2])) * 100) / 100
		case 667, 1639:
			// Oil Temp
			oilTempResistance := binary.BigEndian.Uint16(frame.Data[0:2])
        kelvinTemp := 1 / (A + B * math.Log(float64(oilTempResistance)) + C * math.Pow(math.Log(float64(oilTempResistance)), 3))
			canData.OilTemp = uint16(kelvinTemp - 273.15)
			// Oil Pressure
			oilPressureResistance := float64(binary.BigEndian.Uint16(frame.Data[2:4])) / 819.2
			kPaValue := ((float64(oilPressureResistance) - originalLow) / (originalHigh - originalLow) * (desiredHigh - desiredLow)) + desiredLow
			canData.OilPressure = uint16(math.Round(kPaValue * 0.145038)) // Convert to psi
		}

		jsonData, err := json.Marshal(canData)
		if err != nil {
			log.Println("Json Marshal error (CAN): ", err)
			return
		}
		wsConn.writeToClient(canData.Type, jsonData)
	}
}