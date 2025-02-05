package mazda

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"wt-race-dash/pkg/canUtils"

	"go.einride.tech/can"
)

type CANFrameHandler struct {
	FrameMisc	canUtils.CANFrameMisc
	Frame201	CANFrame201
}

type CANFrame201 struct {
	Type			int			`json:"Type"`
	FrameId		int			`json:"FrameId"`
	Rpm				float64	`json:"Rpm"`
	Tps				float64	`json:"Tps"`
}

func (fh *CANFrameHandler) ProcessCANFrame(frameId uint32, data can.Data, wg sync.WaitGroup, ecuType string, isDatalogging bool) []byte {
	switch (frameId) {
		case 0x69A:
			wg.Add(1)
			go canUtils.DoDatalogging(&isDatalogging, &wg)
			time.Sleep(1 * time.Second)
			isDatalogging = !isDatalogging
			fh.FrameMisc.DataloggingAlert = isDatalogging
			return canUtils.JsonMarshalling(fh.FrameMisc)
		
		case 0x69B:
			fh.FrameMisc.ChangePage = true
			return canUtils.JsonMarshalling(fh.FrameMisc)

		// case 0x69C:

		case 201, 513:
			rpm1 := float64(binary.BigEndian.Uint16(data[0:1]))
			rpm2 := float64(binary.BigEndian.Uint16(data[1:2]))
			fh.Frame201.Rpm = math.Trunc(((256 * rpm1) + rpm2) / 4)
			fh.Frame201.Tps = math.Trunc(float64(binary.BigEndian.Uint16(data[6:7])) / 2)
			return canUtils.JsonMarshalling(fh.Frame201)

		default:
			return nil
	}
}