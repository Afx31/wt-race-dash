# Wills Race Dash

A Raspberry Pi CANBus race dash display for my race car.

Goal: Create a software application to retrieve the can bus data from my race car's Hondata ECU chip and display it on a heads up display. Add additional features like data logging, data sync via cloud, analytics & option to add additional external sensors.

---

A telemetry dash I've built which interfaces with a CANBus network using a Raspberry Pi. The project enables data communication with the vehicle's ECU and supports real-time monitoring and analysis.

Currently, it's configured to work with a Hondata S300, but will look to incorporate individual manufacture protocols and aftermarket ECU protocols in the future. If you request this sooner, drop a GitHub issue.

### The goal
Create a software application to retrieve the CAN Bus data from my race car's Hondata ECU and display it on a heads up display in real-time. Include lap timing functionality, datalogging, and be able to analysis the datalogs from the sesssions/races in a desktop application ([WT - Data Analysis](https://github.com/Afx31/WT-DataAnalysis));


#### Features:
- Datalogging to .csv (can then be read into [WT - Data Analysis](https://github.com/Afx31/WT-DataAnalysis))
- Lap timing
- Multiple predefined pages

#### Configurable settings:
- Datalogging hertz rate
- Warning alert values


---

## Tech specs:

Hardware specs:
- Raspberry Pi 4, 8GB
- Ebay 7" LCD

Software specs:
- Go backend
- HTML/CSS/JavaScript frontend