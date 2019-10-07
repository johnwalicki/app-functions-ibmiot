package transforms

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	sdkTransforms "github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	IBMIoTMQTTHost           = "IBMIoTMQTTHost"
	IBMIoTMQTTPort           = "IBMIoTMQTTPort"
	IBMIoTDeviceType         = "IBMIoTDeviceType"
	IBMIoTDeviceID           = "IBMIoTDeviceID"
	IBMIoTRootCAFilename     = "CaCertPath"
	IBMIoTCertFilename       = "MQTTCert"
	IBMIoTPrivateKeyFilename = "MQTTKey"
	user                     = "use-token-auth"
	topic                    = "iot-2/evt/status/fmt/json"
)

var log logger.LoggingClient

// type certPair struct {
// 	Cert string `json:"cert,omitempty"`
// 	Key  string `json:"key,omitempty"`
// }

// IBMIoTMQTTConfig holds IBM IoT specific information
type IBMIoTMQTTConfig struct {
	MQTTConfig  *sdkTransforms.MqttConfig
	IoTHost     string
	IoTPort     string
	IoTDevice   string
	IoTTopic    string
	KeyCertPair *sdkTransforms.KeyCertPair
}

func getNewClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &http.Client{Timeout: 10 * time.Second, Transport: tr}
}

func getAppSetting(settings map[string]string, name string) string {
	value, ok := settings[name]

	if ok {
		log.Debug(value)
		return value
	}
	log.Error(fmt.Sprintf("ApplicationName application setting %s not found", name))
	return ""

}

// LoadAWSMQTTConfig Loads the mqtt configuration necessary to connect to AWS cloud
func LoadIBMIoTMQTTConfig(sdk *appsdk.AppFunctionsSDK) (*IBMIoTMQTTConfig, error) {
	if sdk == nil {
		return nil, errors.New("Invalid AppFunctionsSDK")
	}

	log = sdk.LoggingClient

	var ioTHost, iotPort, iotDevice, mqttCert, mqttKey, ioTTopic string

	appSettings := sdk.ApplicationSettings()
	if appSettings != nil {
		ioTHost =   getAppSetting(appSettings, IBMIoTMQTTHost)
		iotPort =   getAppSetting(appSettings, IBMIoTMQTTPort)
		iotDevice = getAppSetting(appSettings, IBMIoTThingName)
		mqttCert =  getAppSetting(appSettings, IBMIoTCertFilename)
		mqttKey =   getAppSetting(appSettings, IBMIoTPrivateKeyFilename)
		ioTTopic =  getAppSetting(appSettings, topic)
	} else {
		return nil, errors.New("No application-specific settings found")
	}

	config := IBMIoTMQTTConfig{}

	config.IoTHost = ioTHost
	config.IoTPort = iotPort
	config.IoTDevice = iotDevice
	config.IoTTopic = ioTTopic
	config.MQTTConfig = sdkTransforms.NewMqttConfig()

	pair := &sdkTransforms.KeyCertPair{
		KeyFile:  mqttKey,
		CertFile: mqttCert,
	}

	config.KeyCertPair = pair

	return &config, nil
}

// NewIBMIoTMQTTSender return a mqtt sender capable of sending the event's value to the given MQTT broker
func NewIBMIoTMQTTSender(logging logger.LoggingClient, config *IBMIoTMQTTConfig) *sdkTransforms.MQTTSender {

	logging.Debug(config.IoTTopic)

	port, err := strconv.Atoi(config.IoTPort)
	if err != nil {
		// falling back to default IBM IoT port
		port = 8883
	}

	addressable := models.Addressable{
		Address:   config.IoTHost,
		Port:      port,
		Protocol:  "tls",
		Publisher: "d:org:devicetype:"+config.IoTDevice,
		User:      "",
		Password:  "",
		Topic:     config.IoTTopic,
	}

	mqttSender := sdkTransforms.NewMQTTSender(logging, addressable, config.KeyCertPair, config.MQTTConfig)

	return mqttSender
}
