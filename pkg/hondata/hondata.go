package hondata

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"wt-race-dash/pkg/datalogging"

	"go.einride.tech/can"
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
	Type              int  `json:"Type"`
  CheckEngineLight  bool `json:"CELAlert`
  DataloggingAlert  bool `json:"DataloggingAlert`
  ChangePage        bool `json:"ChangePage`
}

type CANFrame660 struct {
  Type      int     `json:"Type"`
  FrameId   int     `json:"FrameId"`
  Rpm       uint16  `json:"Rpm"`
  Speed     uint16  `json:"Speed"`
  Gear      uint8   `json:"Gear"`
  Voltage   float32 `json:"Voltage"`
}

type CANFrame661 struct {
	Type      int     `json:"Type"`
	FrameId   int     `json:"FrameId"`
	Iat       uint16  `json:"Iat"`
	Ect       uint16  `json:"Ect"`
}

type CANFrame662 struct {
	Type      int     `json:"Type"`
	FrameId   int     `json:"FrameId"`
	Tps       uint16  `json:"Tps"`
	Map       uint16  `json:"Map"`
}

type CANFrame664 struct {
	Type          int     `json:"Type"`
	FrameId       int     `json:"FrameId"`
	LambdaRatio   float64 `json:"LambdaRatio"`
}

type CANFrame667 struct {
	Type         int    `json:"Type"`
	FrameId      int    `json:"FrameId"`
	OilTemp      uint16 `json:"OilTemp"`
	OilPressure  uint16 `json:"OilPressure"`
}

var (
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

func (fh *CANFrameHandler) JsonMarshalling(frameData interface{}) []byte {
	jsonData, err := json.Marshal(frameData)
	if err != nil {
		log.Println("Json Marshal error (CAN): ", err)
		return nil
	}
	return jsonData
}

func (fh *CANFrameHandler) ProcessCANFrame(frameId uint32, data can.Data, wg sync.WaitGroup, isDatalogging bool) []byte {
	switch (frameId) {
		case 67, 103:
			fh.FrameMisc.ChangePage = true
			return fh.JsonMarshalling(fh.FrameMisc)

		case 68, 104:
			wg.Add(1)
			go datalogging.DoDatalogging(&isDatalogging, &wg)
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