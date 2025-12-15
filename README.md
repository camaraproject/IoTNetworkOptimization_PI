# IoT Network Optimization API - Provider Implementation

This repository contains the Provider Implementation for the IoT Network Optimization API. This is a possible implementation of the Transformation Function for the IoT Network Optimization API.
The Provider Implementation is compliant with [r1.1](https://github.com/camaraproject/IoTNetworkOptimization/releases/tag/r1.1) of the IoT Network Optimization API.

## IoT Network Optimization API - Description

This document outlines the implementation details for the IoT Network Optimization API.

The specification of the API is here: https://github.com/camaraproject/IoTNetworkOptimization

The IoT Network Optimization API performs the following tasks:

1) verifies if the API Consumer is allowed to invoke the API for the specific requested applications
2) gathers information about all the identifiers for each device (in particular the NetworkAccessIdentifier for all devices). Information is required to ask the network to enable/disable specific settings for a particular device.
3) as soon as the start date happens (if present, otherwise immediately) gets and stores the current power settings of each requested device
4) applies new settings for each requested device
5) If an end date is present, settings are reverted at that specific moment in time

### Implemented Version

r1.1

### API Functionality

The IoT Network Optimization API supports the following intent:

  - Intent1: I would like to activate power-saving features for my IoT devices during a specified period.

To support the above intent one endpoint is provided:

  - **power-saving:** Configures the power saving features in the network for the provided list of devices.

## How to run service locally

TBD

