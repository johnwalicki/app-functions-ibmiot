package main

import (
	"fmt"
	"os"

	IBMIoTTransforms "github.com/edgexfoundry-holding/app-functions-ibmiot/pkg/transforms"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
)

const (
	serviceKey = "IBMExport"
)

func main() {

	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	// 2) Since our DeviceNameFilter Function requires the list of device names we would
	// like to search for, we'll go ahead and define that now.
	deviceName := []string{"Random-Integer-Generator01"}

	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.

	// Load IBMIoT-specific MQTT configuration from App SDK
	// You can also create IBMIoTMQTTConfig struct yourself
	config, err := IBMIoTTransforms.LoadIBMIoTMQTTConfig(edgexSdk)
	if err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("Failed to load IBMIoT MQTT configurations: %v\n", err))
		os.Exit(-1)
	}

	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceName).FilterByDeviceName,
		IBMIoTTransforms.NewConversion().TransformToIBMIoT,
		printIBMIoTDataToConsole,
		IBMIoTTransforms.NewIBMIoTMQTTSender(edgexSdk.LoggingClient, config).MQTTSend,
	)

	// 5) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}

func printIBMIoTDataToConsole(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	fmt.Println(params[0].(string))

	// Leverage the built in logging service in EdgeX
	edgexcontext.LoggingClient.Debug("Printed to console")

	edgexcontext.Complete([]byte(params[0].(string)))
	return false, nil

}
