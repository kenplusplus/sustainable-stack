package rapl

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"sustainability.collector/pkg/utils"

	"github.com/jaypipes/ghw"
)

const (
	BasePath    = "/sys/class/powercap/"
	RAPLPkgPath = BasePath + "intel-rapl:%d/energy_uj"
	RAPLMemPath = BasePath + "intel-rapl:%d:0/energy_uj"
	PkgMax      = 262143328850
	MemMax      = 65712999613
	//Type: pkg dram
	EnergyType = 2
)

var (
	cpu    *ghw.CPUInfo
	pkgNum int
)

type RAPLEnergy struct{}

func init() {
	var err error
	cpu, err = ghw.CPU()
	if err != nil {
		utils.Sugar.Errorf("get cpu info error: %s\n", err)
	}
	pkgNum = len(cpu.Processors)

}
func (r *RAPLEnergy) Run(quit chan struct{}) {
	// open csv file
	fileName := fmt.Sprintf("energy_result_%s.csv", time.Now().Format("20060102150405"))
	f, err := os.Create(fileName)
	if err != nil {
		utils.Sugar.Panicf("create energy_result.csv error: %s\n", err)
	}
	defer f.Close()

	// read rapl value
	energy := make(chan []string)
	go r.readRAPLHelper(quit, energy, 60*time.Second)

	// store data to csv file
	go func() {
		w := csv.NewWriter(f)
		if err = w.Write([]string{"Package", "Memory"}); err != nil {
			utils.Sugar.Errorf("error writing column header to csv: %s\n", err)
		}
		for v := range energy {
			if err := w.Write(v); err != nil {
				utils.Sugar.Errorf("error writing record to csv: %s\n", err)
			}
			w.Flush()
		}
	}()
	<-quit
}

// readRAPLHelper gets the calculated rapl value based on interval time.
func (r *RAPLEnergy) readRAPLHelper(quit chan struct{}, energy chan []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	pre := &[]uint64{0, 0}
	for {
		select {
		case <-ticker.C:
			energy <- calculateDeltaEnergy(pre)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// ReadCurrentRapl get the current RAPL value.
func (r *RAPLEnergy) ReadCurrentRapl() ([]uint64, error) {
	return getRAPLEnergy(pkgNum)
}

// CalculateEnergy calculate the dynamic energy consumed by the accelerator
// and normalizes the raw data.
func (r *RAPLEnergy) CalculateDynEnergy(idlePower []uint64, preEnergy []uint64, curEnergy []uint64, timeCost float64) []string {

	var dynEnergy = make([]string, 0, EnergyType)

	qatEnergy := calculateEnergy(preEnergy, curEnergy)

	for i := 0; i < EnergyType; i++ {
		tmpEnergy := float64(qatEnergy[i]) - float64(idlePower[i])*timeCost
		dynEnergy = append(dynEnergy, strconv.FormatFloat(tmpEnergy, 'f', 3, 64))
	}

	return dynEnergy
}

// GetIdlePower get the idle power over a period of time, in units of uJ/s
func GetIdlePower(d time.Duration) ([]uint64, error) {
	utils.Sugar.Infof("get idle energy, please wait %0.3f seconds....\n", d.Seconds())
	power := make([]uint64, 0, EnergyType)

	pre, err := getRAPLEnergy(pkgNum)
	if err != nil {
		utils.Sugar.Errorf("get previous info error: %s\n", err)
		return nil, err
	}

	time.Sleep(d)

	cur, err := getRAPLEnergy(pkgNum)
	if err != nil {
		utils.Sugar.Errorf("get current info error: %s\n", err)
		return nil, err
	}

	//energy /*uJ*/
	energy := calculateEnergy(pre, cur)

	//Calculate power /* uJ/s */
	for i := 0; i < EnergyType; i++ {
		power = append(power, energy[i]/uint64(d.Seconds()))
	}
	return power, nil
}
