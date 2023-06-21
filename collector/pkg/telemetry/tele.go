package telemetry

import (
	"time"

	"sustainability.collector/pkg/utils"
)

const (
	off = iota
	on
	addr                   = "6b"
	controlPathTemplate    = "/sys/devices/pci0000:%s/0000:%s:00.0/telemetry/control"
	deviceDataPathTemplate = "/sys/devices/pci0000:%s/0000:%s:00.0/telemetry/device_data"
	interval               = 1 * time.Second
)

// ReadTelemetry gets the telemetry value based on interval time
func ReadTelemetry(stop chan struct{}) (r *ResTelemetry, err error) {

	var res ResTelemetry

	ticker := time.NewTicker(interval)
	//start telemetry
	err = controlTelemetry(addr, on)
	if err != nil {
		return nil, err
	}

	//read device data
	for {
		select {
		case <-stop:
			ticker.Stop()
			//stop telemetry
			err = controlTelemetry(addr, off)
			utils.Sugar.Infoln("receive a signal, stop telemetry")
			return &res, err
		case <-ticker.C:
			data, err := readTelemetry(addr)
			if err != nil {
				return &res, err
			}
			//calculate telemetry value
			if data.latency > 0 {
				res.CntSum++
				res.PciTransSum += data.sampleCnt
				res.LatencySum += data.latency
				res.BwInSum += data.bwIn
				res.BwOutSum += data.bwOut
				res.CprSum += data.cprUtil
				res.DcprSum += data.dcprUtil
			}

		}

	}

}

type ResTelemetry struct {
	CntSum      uint64
	PciTransSum uint64
	LatencySum  uint64
	BwInSum     uint64
	BwOutSum    uint64
	CprSum      uint64
	DcprSum     float64
}
