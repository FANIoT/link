# Link
[![Travis branch](https://img.shields.io/travis/com/I1820/link/master.svg?style=flat-square)](https://travis-ci.com/I1820/link)
[![Go Report](https://goreportcard.com/badge/github.com/I1820/link?style=flat-square)](https://goreportcard.com/report/github.com/I1820/link)
[![Buffalo](https://img.shields.io/badge/powered%20by-buffalo-blue.svg?style=flat-square)](http://gobuffalo.io)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1bdf3a4f0b294e9e92f15211ba894ef4)](https://www.codacy.com/app/i1820/link?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=I1820/link&amp;utm_campaign=Badge_Grade)

## Introduction

Link component of I1820 platfrom. This service collects
raw data from bottom layer (protocols like mqtt, coap and http), stores them into mongo database.
This service also sends data into bottom layer (protocols).

There is two way for setting state in the I1820 platform.
First one is to set a particular asset's state on a specific device with the following JSON:

```json
{
  "value": 10.2,
  "at": "1970-01-01T00:00:00Z"
}
```

The second one is to set a particular device's state with the following JSON:

```json
{
  "asset_name": {
    "value": 10.2,
    "at": "1970-01-01T00:00:00Z"
  }
}
```

These structures are standard data format of I1820 communications.

Beside mentioned ways for setting state, I1820 supports another ways of sending data.
users can send data with the custom format and then decode it with python scripts that are written by their slaves.

## Pipeline
No matter where data is coming, it must have `State` structure to traverse the core pipeline. You can see `State` structure definition in [here](https://github.com/I1820/types/blob/master/state.go).

Pipeline has three main stage as listed below:

- Project Stage: Finds `Porject` of incoming `State` by its `ThingID`.
- Decode Stage: Decodes `Raw` into `Value` by finding its type.
- Insert Stage: Inserts `State` into mongo database.

## Inner vs Outer Broker
Link component publishes data for inner component on MQTT, sometimes
inner and outer brokers are different so we have two following configuration
for brokers:

- `SYS_BROKER_URL # internal broker`
- `USR_BROKER_URL # outer broker`

## MQTT Protocol
For changing device state using mqtt protocol you can use following topic:

- `things/{thing_id}/state`

You must send standard JSON with the structure mentioned earlier.

## API

Link parses and decodes each incoming data then stores them into Mongo database.
For providing a way for other components to have data, it publishes data into the following topics:

- `i1820/projects/{project_id}/things/{thing_id}/assets/{asset_name}/state`
