# Hallway bus monitor

I needed a little program to monitor the buss arrival times in my hallway using a raspberry pi, monitor and resrobot api.

## Installation

    go install github.com/hajhatten/hallmonitor

## Usage
    RESROBOTAPIKEY=<APIKEY> while true; do clear; hallmonitor; sleep 300; done