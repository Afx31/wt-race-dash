package mazda

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
	FrameMisc	CANFrameMisc
	Frame201	CANFrame201
}

type CANFrameMisc struct {
	Type              int  `json:"Type"`
  CheckEngineLight  bool `json:"CELAlert`
  DataloggingAlert  bool `json:"DataloggingAlert`
  ChangePage        bool `json:"ChangePage`
}

type CANFrame201 struct {
	Type			int			`json:"Type"`
	FrameId		int			`json:"FrameId"`
	Rpm				float64	`json:"Rpm"`
	Tps				float64	`json:"Tps"`
}

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
	
	case 201, 513:
		rpm1 := float64(binary.BigEndian.Uint16(data[0:1]))
		rpm2 := float64(binary.BigEndian.Uint16(data[1:2]))
		fh.Frame201.Rpm = math.Trunc(((256 * rpm1) + rpm2) / 4)
		fh.Frame201.Tps = math.Trunc(float64(binary.BigEndian.Uint16(data[6:7])) / 2)
		return fh.JsonMarshalling(fh.Frame201)

	default:
		return nil
	}
}