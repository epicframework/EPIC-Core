# EPIC-Core

The Extensible Package Integration Controller is an IoT framework which utilizes a custom-built, light-weight key-value store and a set of core applications to provide users the capability to seamlessly integrate IoT devices and expand the base functionality of the framework. The core of the framework includes the manager application, a web-based user interface for deploying new packages and managing devices and a generic device manager for connecting WiFi enabled IoT devices based on the Mozilla Web of Things standard, as well as a set of python bindings for integrating these applications and any packages with Connective, the key-value store and main communication functionality of the framework.

## Manager

- Initializes the core applications of the framework. 
- Monitors device heartbeats (Compeleted) and validates connective transaction receipts (In Progress).
- Deploys new packages uploaded through the framework web interface using Docker
- Performs failure recovery and prevention operations

## Server

- Displays all existing packages and devices in a simple user interface
- Allows for the deployment of new packages through Connective to Manager
- Provides IoT device control through Device Manager and an API based on the Mozilla Web of Things standard

## Device Manager

- Locates and initiates a connection with compatible EPIC IoT smart devices
- Receives devices specification and capabilities
- Exposes device capabilities to Server and integrated packages through Connective

## Connective

- Allows for the creation of queues and directories
- Creates transaction receipts to be validated by Manager
- Facilitates heartbeat generation and monitoring

## Packages

Packages are utilized to expand the functionality of the core framework. Through Connective, packages are capable of utilizing all information exposed and generated through the core applications as well as generating their own data and exposing different classes of devices. Examples of packages are specialized drivers for handling Bluetooth and ZigBee devices and conforming their behaviour to the Mozilla Web of Things-based implementation used for WiFi devices, addtional user interfaces or utilities, or the implementation of data management tools.
