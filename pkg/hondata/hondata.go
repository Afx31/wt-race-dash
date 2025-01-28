package hondata

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"wt-race-dash/pkg/canUtils"

	"go.einride.tech/can"
)

type CANFrameHandler struct {
	FrameMisc canUtils.CANFrameMisc
	Frame660  CANFrame660
	Frame661  CANFrame661
	Frame662  CANFrame662
  Frame663  CANFrame663
	Frame664  CANFrame664
  Frame665  CANFrame665
  Frame666  CANFrame666
	Frame667  CANFrame667
  //Frame668  CANFrame668
  Frame669  CANFrame669
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
  Mil       uint8   `json:"Mil"`
  Vts       uint8   `json:"Vts"`
  Cl        uint8   `json:"Cl"`
}

type CANFrame662 struct {
	Type      int    `json:"Type"`
	FrameId   int    `json:"FrameId"`
	Tps       uint16 `json:"Tps"`
	Map       uint16 `json:"Map"`
}

type CANFrame663 struct {
	Type      int    `json:"Type"`
	FrameId   int    `json:"FrameId"`
	Inj       uint16 `json:"Inj"`
	Ign       uint16 `json:"Ign"`
}

type CANFrame664 struct {
	Type          int     `json:"Type"`
	FrameId       int     `json:"FrameId"`
	LambdaRatio   float64 `json:"LambdaRatio"`
}

type CANFrame665 struct {
  Type          int   `json:"Type"`
  FrameId       int   `json:"FrameId"`
  KnockCounter  int   `json:"KnockCounter"`
}

type CANFrame666 struct {
  Type            int     `json:"Type"`
  FrameId         int     `json:"FrameId"`
  TargetCamAngle  float64 `json:"TargetCamAngle"`
  ActualCamAngle  float64 `json:"ActualCamAngle"`
}

type CANFrame667 struct {
	Type         int    `json:"Type"`
	FrameId      int    `json:"FrameId"`
	OilTemp      uint16 `json:"OilTemp"`
	OilPressure  uint16 `json:"OilPressure"`
  // Analog 3
  // Analog 4
}

// TODO: Future
// type CANFrame668 struct {
  // Analog 1
  // Analog 2
  // Analog 3
  // Analog 4
// }

type CANFrame669 struct {
  Type            int     `json:"Type"`
  FrameId         int     `json:"FrameId"`
  EthanolInput1		int     `json:"EthanolInput1"`
  EthanolInput2   float64  `json:"EthanolInput2"`
  EthanolInput3		uint16  `json:"EthanolInput3"`
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

func (fh *CANFrameHandler) ProcessCANFrame(frameId uint32, data can.Data, wg sync.WaitGroup, ecuType string, isDatalogging bool) []byte {
	switch (frameId) {
		case 67, 103:
			fh.FrameMisc.ChangePage = true
			return canUtils.JsonMarshalling(fh.FrameMisc)

		case 68, 104:
			wg.Add(1)
			go canUtils.DoDatalogging(&isDatalogging, &wg)
			time.Sleep(1 * time.Second)
			isDatalogging = !isDatalogging
			fh.FrameMisc.DataloggingAlert = isDatalogging
			return canUtils.JsonMarshalling(fh.FrameMisc)

		// case 69, 105:
			// TODO: Will need to read in whatever the value is and perform the bool conversion below
			//fh.FrameMisc.CheckEngineLight = !fh.FrameMisc.CheckEngineLight
			// return canUtils.JsonMarshalling(fh.FrameMisc)

		case 660, 1632:
			fh.Frame660.Rpm = binary.BigEndian.Uint16(data[0:2])
			fh.Frame660.Speed = binary.BigEndian.Uint16(data[2:4])
			fh.Frame660.Gear = data[4]
			fh.Frame660.Voltage = float32(data[5]) / 10.0
			return canUtils.JsonMarshalling(fh.Frame660)
		
		case 661, 1633:
			fh.Frame661.Iat = binary.BigEndian.Uint16(data[0:2])
			fh.Frame661.Ect = binary.BigEndian.Uint16(data[2:4])
      if (ecuType == "kpro") {
        fh.Frame661.Mil = uint8(binary.BigEndian.Uint16(data[4:5]))
        fh.Frame661.Vts = uint8(binary.BigEndian.Uint16(data[5:6]))
        fh.Frame661.Cl = uint8(binary.BigEndian.Uint16(data[6:7]))
      }
			return canUtils.JsonMarshalling(fh.Frame661)
		
		case 662, 1634:
			fh.Frame662.Tps = binary.BigEndian.Uint16(data[0:2])
				if fh.Frame662.Tps == 65535 { fh.Frame662.Tps = 0	}
			fh.Frame662.Map = binary.BigEndian.Uint16(data[2:4]) / 10
			return canUtils.JsonMarshalling(fh.Frame662)
		
    case 663, 1635:
			fh.Frame663.Inj = binary.BigEndian.Uint16(data[0:2]) / 1000
			fh.Frame663.Ign = binary.BigEndian.Uint16(data[2:4])
			return canUtils.JsonMarshalling(fh.Frame663)

		case 664, 1636:
			fh.Frame664.LambdaRatio = math.Round(float64(32768.0) / float64(binary.BigEndian.Uint16(data[0:2])) * 100) / 100
			return canUtils.JsonMarshalling(fh.Frame664)
		
    // K-Pro only
    case 665, 1637:
			if (ecuType == "kpro") {
      	fh.Frame665.KnockCounter = int(binary.BigEndian.Uint16(data[0:2]))
      	return canUtils.JsonMarshalling(fh.Frame665)
			} else {
				return nil
			}
			

    // K-Pro only
    case 666, 1638:
			if (ecuType == "kpro") {
      	fh.Frame666.TargetCamAngle = float64(binary.BigEndian.Uint16(data[0:2]))
      	fh.Frame666.ActualCamAngle = float64(binary.BigEndian.Uint16(data[2:4]))
      	return canUtils.JsonMarshalling(fh.Frame666)
			} else {
				return nil
			}

		case 667, 1639:
			// Oil Temp
			oilTempResistance := binary.BigEndian.Uint16(data[0:2])
				kelvinTemp := 1 / (A + B * math.Log(float64(oilTempResistance)) + C * math.Pow(math.Log(float64(oilTempResistance)), 3))
			fh.Frame667.OilTemp = uint16(kelvinTemp - 273.15)
			// Oil Pressure
			oilPressureResistance := float64(binary.BigEndian.Uint16(data[2:4])) / 819.2
			kPaValue := ((float64(oilPressureResistance) - originalLow) / (originalHigh - originalLow) * (desiredHigh - desiredLow)) + desiredLow
			fh.Frame667.OilPressure = uint16(math.Round(kPaValue * 0.145038)) // Convert to psi
			return canUtils.JsonMarshalling(fh.Frame667)

    // TODO: Future
    // case 668, 1640:

    case 669, 1641:
			fh.Frame669.EthanolInput1 = int(binary.BigEndian.Uint16(data[0:1])) // Frequency

			if (ecuType == "s300") {
      	fh.Frame669.EthanolInput2 = float64(binary.BigEndian.Uint16(data[1:2])) * 2.56 // Duty
      	fh.Frame669.EthanolInput3 = binary.BigEndian.Uint16(data[2:3]) // Content
			} else if (ecuType == "kpro") {
      	fh.Frame669.EthanolInput2 = float64(binary.BigEndian.Uint16(data[1:2])) // Ethanol Content
      	fh.Frame669.EthanolInput3 = binary.BigEndian.Uint16(data[2:4]) // Fuel Temperature
			}
      return canUtils.JsonMarshalling(fh.Frame669)

		default:
			return nil
	}
}