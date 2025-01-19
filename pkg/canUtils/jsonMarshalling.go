package canUtils

import (
	"encoding/json"
	"log"
)

func JsonMarshalling(frameData interface{}) []byte {
	jsonData, err := json.Marshal(frameData)
	if err != nil {
		log.Println("Json Marshal error (CAN): ", err)
		return nil
	}
	return jsonData
}