package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"

	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/viper"
)

var logger *log.Logger

func init() {
	// Set up logging
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Load configuration
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatalf("error reading config file: %v", err)
	}

	deviceName := viper.GetString("AppSettings.DeviceName")
	logger.Println("Hello, World!")
	logger.Printf("DeviceName: %s", deviceName)

	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		logger.Fatalf("error finding devices: %v", err)
	}

	if len(devices) < 1 {
		logger.Println("No devices were found on this machine")
		return
	}

	for i, device := range devices {
		logger.Printf("%d) %s - %s", i+1, device.Name, device.Description)
	}

	var network *pcap.Interface
	for _, device := range devices {
		if strings.Contains(device.Description, deviceName) {
			network = &device
			break
		}
	}

	if network != nil {
		handlePacketCapture(*network)
	} else {
		logger.Println("No devices were found for capture")
	}
}

func handlePacketCapture(device pcap.Interface) {
	logger.Println("Capture Started...")

	handle, err := pcap.OpenLive(device.Name, 1600, true, pcap.BlockForever)
	if err != nil {
		logger.Fatalf("error opening device: %v", err)
	}
	defer handle.Close()

	// Set the filter for UDP port 5060
	err = handle.SetBPFFilter("udp port 5060")
	if err != nil {
		logger.Fatalf("error setting BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	go func() {
		for packet := range packetSource.Packets() {
			if isSIPPacket(packet) {
				packetJSON, err := packetToJSON(packet)
				if err != nil {
					logger.Printf("error converting packet to JSON: %v", err)
					continue
				}
				//logger.Println("packetJSON = %s", len(packetJSON))
				logger.Println("SIP Packet Captured: %s", packetJSON)
			}
		}
	}()

	// Signal handling for graceful shutdown
	logger.Println("Running indefinitely. Press Ctrl+C to stop.")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	logger.Println("Shutting down...")
}

func isSIPPacket(packet gopacket.Packet) bool {
	// Check if the packet is a UDP packet
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return false
	}

	// Optionally, you can inspect UDP payload here to confirm it's SIP (e.g., check for SIP headers)
	// For simplicity, we'll assume any UDP packet on port 5060 is SIP, based on the BPF filter
	return true
}

func packetToJSON(packet gopacket.Packet) (string, error) {
	packetInfo := map[string]interface{}{
		"timestamp": packet.Metadata().Timestamp,
		"length":    packet.Metadata().Length,
	}

	// Decode layers
	layerInfo := map[string]interface{}{}
	for _, layer := range packet.Layers() {
		layerType := layer.LayerType().String()
		layerData := layer.LayerPayload() // Get the raw payload of the layer

		// Convert payload to string
		//layerDataStr := string(layerData)
		//logger.Println(layerDataStr)
		if layerType == "UDP" {

			layerDataStr := string(layerData)
			//logger.Println("layerData = %s", len(layerData))
			//logger.Println("layerDataStr = %s", len(layerDataStr))
			//logger.Println(layerDataStr)

			if len(layerData) <= 2 {
				continue
			}
			//logger.Println(layerDataStr)
			layerInfo[layerType] = map[string]interface{}{
				"payload": layerDataStr, // Convert payload to string
				//"hex":     toHex(layerData), // hexadecimal representation
			}
			packetInfo["layers"] = layerInfo

			//Parse SIP message
			sipMessage := parseSIPMessage(layerDataStr)
			packetInfo["message"] = sipMessage
			break
		}
	}

	if packetInfo["message"] != nil {
		// Convert to JSON
		packetJSON, err := json.MarshalIndent(packetInfo, "", "  ")
		if err != nil {
			return "", err
		}
		//logger.Println("aaaa")
		return string(packetJSON), nil
	} else {
		//logger.Println("bbb")
		return "", nil
	}

}

// Convert bytes to a hex string representation
func toHex(data []byte) string {
	return hex.EncodeToString(data)
}

func isValidUTF8(data []byte) bool {
	for _, b := range data {
		if b > 127 {
			return false
		}
	}
	return true
}

func parseSIPMessage(sipMessage string) map[string]interface{} {
	lines := strings.Split(sipMessage, "\r\n")
	sipHeaders := map[string]string{}

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				sipHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return map[string]interface{}{
		"headers": sipHeaders,
		"body":    "", // You can parse the body if needed
	}
}

func isOnlyLineBreaks(data []byte) bool {
	for _, b := range data {
		if b != '\r' && b != '\n' {
			return false
		}
	}
	return true
}
