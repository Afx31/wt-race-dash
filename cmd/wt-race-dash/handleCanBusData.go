package main

import (
	"context"
	"fmt"
	"sync"

	"wt-race-dash/pkg/canUtils"
	"wt-race-dash/pkg/hondata"
	"wt-race-dash/pkg/mazda"

	"go.einride.tech/can"
	"go.einride.tech/can/pkg/socketcan"
)

type CANInterface interface {
	ProcessCANFrame(frameId uint32, data can.Data, wg sync.WaitGroup, ecuType string, isDatalogging *bool) []byte
}

var (
	isDatalogging = false
	canInterface CANInterface
)

func (wsConn *MySocket) HandleCanBusData() {
	// ---------- CANBus data ----------
	canConn, err := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
	if err != nil {
		// "Failed to connect to CAN channel"
		fmt.Println("SocketCAN Connection Error: ", err)
		fmt.Println("==========================================")
	}
	defer canConn.Close()
	canRecv := socketcan.NewReceiver(canConn)

	switch (appSettings.Car) {
	case "honda":
		canInterface = &hondata.CANFrameHandler{
			FrameMisc: canUtils.CANFrameMisc{ Type: 5 },
			Frame660: hondata.CANFrame660{ Type: 1, FrameId: 660 },
			Frame661: hondata.CANFrame661{ Type: 1, FrameId: 661 },
			Frame662: hondata.CANFrame662{ Type: 1, FrameId: 662 },
      Frame663: hondata.CANFrame663{ Type: 1, FrameId: 663 },
			Frame664: hondata.CANFrame664{ Type: 1, FrameId: 664 },
			Frame667: hondata.CANFrame667{ Type: 1, FrameId: 667 },
      Frame669: hondata.CANFrame669{ Type: 1, FrameId: 669 },
		}

    if (appSettings.Ecu == "kpro") {
      canInterface.(*hondata.CANFrameHandler).Frame665 = hondata.CANFrame665{ Type: 1, FrameId: 665 }
      canInterface.(*hondata.CANFrameHandler).Frame666 = hondata.CANFrame666{ Type: 1, FrameId: 666 }
    }
	case "mazda":
		canInterface = &mazda.CANFrameHandler{
			FrameMisc: canUtils.CANFrameMisc{ Type: 5 },
			Frame201: mazda.CANFrame201{ Type: 1, FrameId: 201 },
		}
	}

	for canRecv.Receive() {
		frame := canRecv.Frame()

		jsonData := canInterface.ProcessCANFrame(frame.ID, frame.Data, wg, appSettings.Ecu, &isDatalogging)
		
		if ((frame.ID == 67 || frame.ID == 103) && canInterface.(*hondata.CANFrameHandler).FrameMisc.ChangePage) {
			canInterface.(*hondata.CANFrameHandler).FrameMisc.ChangePage = false
		}

		if jsonData != nil {
			wsConn.writeToClient(int8(frame.ID), jsonData)
		}
  }
}