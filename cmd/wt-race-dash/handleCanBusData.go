package main

import (
	"context"
	// "encoding/binary"
	// "encoding/json"
	"fmt"
	// "log"
	// "math"
	// "os/exec"
	// "sync"
	// "time"

	// "go.einride.tech/can"
	"wt-race-dash/pkg/hondata"

	"go.einride.tech/can/pkg/socketcan"
	// "wt-race-dash/pkg/datalogging"
)

var (
	isDatalogging = false
	tempStruct interface{}
)

func (wsConn *MySocket) HandleCanBusData() {
	// ---------- CANBus data ----------
	canConn, err := socketcan.DialContext(context.Background(), "can", appSettings.CanChannel)
	if err != nil {
		// "Failed to connect to CAN channel"
		fmt.Println("SocketCAN Connection Error: ", err)
		fmt.Println("==========================================")
		fmt.Printf("SocketCAN Connection Error: %v", err)
		fmt.Println("==========================================")
	}
	defer canConn.Close()
	canRecv := socketcan.NewReceiver(canConn)

	switch (appSettings.CarOrEcu) {
	case "hondata":
		tempStruct = &hondata.CANFrameHandler{
			FrameMisc: hondata.CANFrameMisc{ Type: 5 },
			Frame660: hondata.CANFrame660{ Type: 1 },
			Frame661: hondata.CANFrame661{ Type: 1 },
			Frame662: hondata.CANFrame662{ Type: 1 },
			Frame664: hondata.CANFrame664{ Type: 1 },
			Frame667: hondata.CANFrame667{ Type: 1 },	
		}
		break;
	}

	for canRecv.Receive() {
		frame := canRecv.Frame()
		
		// OLD
		// jsonData := canFrameHandler.ProcessCANFrame(frame.ID, frame.Data)

		// Hacky but she'll be right for now
		// if ((frame.ID == 67 || frame.ID == 103) && canFrameHandler.FrameMisc.ChangePage) {
		// 	canFrameHandler.FrameMisc.ChangePage = false
		// }

		// if jsonData != nil {
		// 	wsConn.writeToClient(int8(frame.ID), jsonData)
		// }

		

		// NEW
		//jsonData := tempStruct.ProcessCANFrame(frame.ID, frame.Data, wg, isDatalogging)
		
		// Asserts the actual type into the interface
		// TODO: error handle the assert
		jsonData := tempStruct.(*hondata.CANFrameHandler).ProcessCANFrame(frame.ID, frame.Data, wg, isDatalogging)

		if ((frame.ID == 67 || frame.ID == 103) && tempStruct.(*hondata.CANFrameHandler).FrameMisc.ChangePage) {
			tempStruct.(*hondata.CANFrameHandler).FrameMisc.ChangePage = false
		}

		if jsonData != nil {
			wsConn.writeToClient(int8(frame.ID), jsonData)
		}
  }
}