package telemetry

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"sustainability.collector/pkg/utils"
)

// controlTelemetry Control telemetry on/off
func controlTelemetry(addr string, mode int) error {

	controlPath := fmt.Sprintf(controlPathTemplate, addr, addr)

	controlFile, err := os.OpenFile(controlPath, os.O_WRONLY, 0644)
	if err != nil {
		utils.Sugar.Panicf("failed to open telemetry control file: %s \n", err)
	}
	defer controlFile.Close()

	_, err = controlFile.WriteString(strconv.Itoa(mode))
	if err != nil {
		utils.Sugar.Errorf("failed to control:%d telemetry: %s \n", mode, err)
		return err
	}

	return nil
}

// readTelemetry Read telemetry data from the device_data file
func readTelemetry(addr string) (*telemetryData, error) {

	deviceDataPath := fmt.Sprintf(deviceDataPathTemplate, addr, addr)

	//open device data file
	deviceDataFile, err := os.OpenFile(deviceDataPath, os.O_RDONLY, 0666)
	if err != nil {
		utils.Sugar.Panicf("failed to open davice_data file:%s \n", err)

	}
	defer deviceDataFile.Close()

	data, err := io.ReadAll(deviceDataFile)
	if err != nil {
		utils.Sugar.Panicf("failed to read device_data:%s \n", err)
	}

	out := strings.Fields(string(data))

	output := make(map[string]uint64)

	for i := 0; i < len(out)-1; i += 2 {
		key := out[i]
		value, _ := strconv.ParseUint(out[i+1], 10, 64)
		output[key] = value
	}

	var dcprUtil float64

	decompress := output["util_dcpr0"] + output["util_dcpr1"] + output["util_dcpr2"]

	if decompress > 0 {
		dcprUtil = float64(decompress) / 3
	}

	td := &telemetryData{
		sampleCnt:   output["sample_cnt"],
		pciTransCnt: output["pci_trans_cnt"],
		latency:     output["lat_acc_avg"],
		bwIn:        output["bw_in"],
		bwOut:       output["bw_out"],
		cprUtil:     output["util_cpr0"],
		dcprUtil:    dcprUtil,
	}

	return td, nil
}

type telemetryData struct {
	sampleCnt   uint64
	pciTransCnt uint64
	latency     uint64
	bwIn        uint64
	bwOut       uint64
	cprUtil     uint64
	dcprUtil    float64
}
