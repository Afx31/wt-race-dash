package canUtils

type CANFrameMisc struct {
	Type							int		`json:"Type"`
	CheckEngineLight	bool 	`json:"CheckEngineLight"`
	DataloggingAlert	bool	`json:"DataloggingAlert"`
	ChangePage				bool	`json:"ChangePage"`
}