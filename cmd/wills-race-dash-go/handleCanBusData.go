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

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

type CANFrameHandler struct {
	FrameMisc CANFrameMisc
	Frame660  CANFrame660
	Frame661  CANFrame661
	Frame662  CANFrame662
	Frame664  CANFrame664
	Frame667  CANFrame667
}

type CANFrameMisc struct {
	Type 							int  `json:"Type"`
	CheckEngineLight	bool `json:"CELAlert`
	DataloggingAlert 	bool `json:"DataloggingAlert`
  ChangePage        bool `json:"ChangePage`
}

type CANFrame660 struct {
	Type    	int     `json:"Type"`
	FrameId 	int			`json:"FrameId"`
	Rpm     	uint16  `json:"Rpm"`
	Speed   	uint16  `json:"Speed"`
	Gear    	uint8   `json:"Gear"`
	Voltage 	float32 `json:"Voltage"`
}

type CANFrame661 struct {
	Type 			int  		`json:"Type"`
	FrameId 	int  		`json:"FrameId"`
	Iat  			uint16 	`json:"Iat"`
	Ect  			uint16 	`json:"Ect"`
}

type CANFrame662 struct {
	Type 			int    	`json:"Type"`
	FrameId 	int    	`json:"FrameId"`
	Tps  			uint16 	`json:"Tps"`
	Map  			uint16 	`json:"Map"`
}

type CANFrame664 struct {
	Type        	int     `json:"Type"`
	FrameId       int     `json:"FrameId"`
	LambdaRatio 	float64 `json:"LambdaRatio"`
}

type CANFrame667 struct {
	Type         int    `json:"Type"`
	FrameId      int    `json:"FrameId"`
	OilTemp      uint16 `json:"OilTemp"`
	OilPressure  uint16 `json:"OilPressure"`
}


var (
	canFrameHandler = &CANFrameHandler{
		FrameMisc: CANFrameMisc{ Type: 5 },
		Frame660: CANFrame660{ Type: 1 },
		Frame661: CANFrame661{ Type: 1 },
		Frame662: CANFrame662{ Type: 1 },
		Frame664: CANFrame664{ Type: 1 },
		Frame667: CANFrame667{ Type: 1 },
	}
	isDatalogging = false

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


func (fh *CANFrameHandler) JsonMarshalling(frameData interface{}) []byte {
	jsonData, err := json.Marshal(frameData)
	if err != nil {
		log.Println("Json Marshal error (CAN): ", err)
		return nil
	}
	return jsonData
}

func (fh *CANFrameHandler) ProcessCANFrame(frameId uint32, data can.Data) []byte {
	switch (frameId) {
		case 67, 103:
			fh.FrameMisc.ChangePage = true
			return fh.JsonMarshalling(fh.FrameMisc)

		case 68, 104:
			wg.Add(1)
			go doDatalogging(&isDatalogging, &wg)
			time.Sleep(1 * time.Second)
			isDatalogging = !isDatalogging
			fh.FrameMisc.DataloggingAlert = isDatalogging
			return fh.JsonMarshalling(fh.FrameMisc)

		// case 69, 105:
			// TODO: Will need to read in whatever the value is and perform the bool conversion below
			//fh.FrameMisc.CheckEngineLight = !fh.FrameMisc.CheckEngineLight
			// return fh.JsonMarshalling(fh.FrameMisc)

		case 660, 1632:
			fh.Frame660.FrameId = 660
			fh.Frame660.Rpm = binary.BigEndian.Uint16(data[0:2])
			fh.Frame660.Speed = binary.BigEndian.Uint16(data[2:4])
			fh.Frame660.Gear = data[4]
			fh.Frame660.Voltage = float32(data[5]) / 10.0
			return fh.JsonMarshalling(fh.Frame660)
		
		case 661, 1633:
			fh.Frame661.FrameId = 661
			fh.Frame661.Iat = binary.BigEndian.Uint16(data[0:2])
			fh.Frame661.Ect = binary.BigEndian.Uint16(data[2:4])
			return fh.JsonMarshalling(fh.Frame661)
		
		case 662, 1634:
			fh.Frame662.FrameId = 662
			fh.Frame662.Tps = binary.BigEndian.Uint16(data[0:2])
				if fh.Frame662.Tps == 65535 { fh.Frame662.Tps = 0	}
			fh.Frame662.Map = binary.BigEndian.Uint16(data[2:4]) / 10
			return fh.JsonMarshalling(fh.Frame662)
		
		case 664, 1636:
			fh.Frame664.FrameId = 664
			fh.Frame664.LambdaRatio = math.Round(float64(32768.0) / float64(binary.BigEndian.Uint16(data[0:2])) * 100) / 100
			return fh.JsonMarshalling(fh.Frame664)
		
		case 667, 1639:
			fh.Frame667.FrameId = 667
			// Oil Temp
			oilTempResistance := binary.BigEndian.Uint16(data[0:2])
				kelvinTemp := 1 / (A + B * math.Log(float64(oilTempResistance)) + C * math.Pow(math.Log(float64(oilTempResistance)), 3))
			fh.Frame667.OilTemp = uint16(kelvinTemp - 273.15)
			// Oil Pressure
			oilPressureResistance := float64(binary.BigEndian.Uint16(data[2:4])) / 819.2
			kPaValue := ((float64(oilPressureResistance) - originalLow) / (originalHigh - originalLow) * (desiredHigh - desiredLow)) + desiredLow
			fh.Frame667.OilPressure = uint16(math.Round(kPaValue * 0.145038)) // Convert to psi
			return fh.JsonMarshalling(fh.Frame667)

		default:
			return nil
	}
}

func (wsConn *MySocket) HandleCanBusData() {
	// ---------- CANBus data ----------
	canConn, _ := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
	defer canConn.Close()
	canRecv := socketcan.NewReceiver(canConn)

	for canRecv.Receive() {
		frame := canRecv.Frame()
		jsonData := canFrameHandler.ProcessCANFrame(frame.ID, frame.Data)

		// Hacky but she'll be right for now
		if ((frame.ID == 67 || frame.ID == 103) && canFrameHandler.FrameMisc.ChangePage) {
			canFrameHandler.FrameMisc.ChangePage = false
		}

		if jsonData != nil {
			wsConn.writeToClient(int8(frame.ID), jsonData)
		}
	}
}

// type CanData struct {
// 	Type             int8
// 	Rpm              uint16
// 	Speed            uint16
// 	Gear             uint8
// 	Voltage          float32
// 	Iat              uint16
// 	Ect              uint16
// 	Tps              uint16
// 	Map              uint16
// 	LambdaRatio      float64
// 	OilTemp          uint16
// 	OilPressure      uint16
//   DataloggingAlert bool
// }

// func (wsConn *MySocket) HandleCanBusData() {
// 	// ---------- CANBus data ----------
// 	canConn, _ := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
// 	defer canConn.Close()
// 	canRecv := socketcan.NewReceiver(canConn)
// 	isDatalogging := false
// 	canData := CanData{Type: 1}

// 	for canRecv.Receive() {
// 		frame := canRecv.Frame()

// 		switch frame.ID {
// 		case 69, 105:
// 			wg.Add(1)
// 			go doDatalogging(&isDatalogging, &wg)
// 			time.Sleep(1 * time.Second)
// 			isDatalogging = !isDatalogging
// 			canData.DataloggingAlert = isDatalogging
		
// 		case 660, 1632:
// 			canData.Rpm = binary.BigEndian.Uint16(frame.Data[0:2])
// 			canData.Speed = binary.BigEndian.Uint16(frame.Data[2:4])
// 			canData.Gear = frame.Data[4]
// 			canData.Voltage = float32(frame.Data[5]) / 10.0
		
// 		case 661, 1633:
// 			canData.Iat = binary.BigEndian.Uint16(frame.Data[0:2])
// 			canData.Ect = binary.BigEndian.Uint16(frame.Data[2:4])
		
// 		case 662, 1634:
// 			canData.Tps = binary.BigEndian.Uint16(frame.Data[0:2])
// 				if canData.Tps == 65535 { canData.Tps = 0	}
// 			canData.Map = binary.BigEndian.Uint16(frame.Data[2:4]) / 10
		
// 		case 664, 1636:
// 			canData.LambdaRatio = math.Round(float64(32768.0) / float64(binary.BigEndian.Uint16(frame.Data[0:2])) * 100) / 100
		
// 		case 667, 1639:
// 			// Oil Temp
// 			oilTempResistance := binary.BigEndian.Uint16(frame.Data[0:2])
//         kelvinTemp := 1 / (A + B * math.Log(float64(oilTempResistance)) + C * math.Pow(math.Log(float64(oilTempResistance)), 3))
// 			canData.OilTemp = uint16(kelvinTemp - 273.15)
// 			// Oil Pressure
// 			oilPressureResistance := float64(binary.BigEndian.Uint16(frame.Data[2:4])) / 819.2
// 			kPaValue := ((float64(oilPressureResistance) - originalLow) / (originalHigh - originalLow) * (desiredHigh - desiredLow)) + desiredLow
// 			canData.OilPressure = uint16(math.Round(kPaValue * 0.145038)) // Convert to psi
// 		}

// 		jsonData, err := json.Marshal(canData)
// 		if err != nil {
// 			log.Println("Json Marshal error (CAN): ", err)
// 			return
// 		}
// 		wsConn.writeToClient(canData.Type, jsonData)
// 	}
// }