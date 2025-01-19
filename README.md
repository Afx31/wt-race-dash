# WillTech - Race Dash

A telemetry dash I've built which interfaces with a CANBus network using a Raspberry Pi. The project enables data communication with the vehicle's ECU and supports real-time monitoring and analysis.

Currently, it's configured to work with a Hondata S300, but will look to incorporate individual manufacture protocols and aftermarket ECU protocols in the future. If you request this sooner, drop a GitHub issue.

### Goal
Create a software application to retrieve the CAN Bus data from my race car's Hondata ECU and display it on a heads up display in real-time. Include lap timing functionality, datalogging, and be able to analysis the datalogs from the sesssions/races in a desktop application ([WT - Data Analysis](https://github.com/Afx31/WT-DataAnalysis));


## Features/settings

Features:
- Datalogging to .csv (can then be read into [WT - Data Analysis](https://github.com/Afx31/WT-DataAnalysis))
- Lap timing
- Multiple predefined pages

Configurable settings:
- Datalogging hertz rate
- Warning alert values

Cars configured for:
- Hondata S300/KPro
- Mazda (Currently reverse engineering a 2011 Mazda 6 GH, shows similar values to Mazda3 and RX-8)

## Tech specs:

Hardware specs:
- Raspberry Pi 4, 8GB
- Ebay 7" LCD

Software specs:
- Go backend
- HTML/CSS/JavaScript frontend
