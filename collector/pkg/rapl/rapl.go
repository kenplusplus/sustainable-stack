package rapl

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"sustainability.amx/pkg/utils"
)

const (
	BasePath    = "/sys/class/powercap/"
	RAPLPkgPath = BasePath + "intel-rapl:%d/energy_uj"
	RAPLMemPath = BasePath + "intel-rapl:%d:0/energy_uj"
	PkgMax      = 262143328850
	MemMax      = 65712999613
)

type RAPLPower struct{}

func (r *RAPLPower) Run(quit chan struct{}) {
	// open csv file
	fileName := fmt.Sprintf("power_result_%s.csv", time.Now().Format("20060102150405"))
	f, err := os.Create(fileName)
	if err != nil {
		utils.Sugar.Panicf("create power_result.csv error: %s\n", err)
	}
	defer f.Close()

	// read rapl value
	power := make(chan []string)
	go r.readRAPLHelper(quit, power, 60*time.Second)

	// store data to csv file
	go func() {
		w := csv.NewWriter(f)
		w.Write([]string{"Package", "Memory"})
		for v := range power {
			if err := w.Write(v); err != nil {
				utils.Sugar.Errorf("error writing record to csv: %s\n", err)
			}
			w.Flush()
		}
	}()
	<-quit
}

// readRAPLHelper gets the calculated rapl value based on interval time.
func (r *RAPLPower) readRAPLHelper(quit chan struct{}, power chan []string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	pre := &[]uint64{0, 0}
	for {
		select {
		case <-ticker.C:
			power <- calculateDeltaPower(pre)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// calculateDeltaPower computes the rapl delta value compared with previous value,
// and normalizes the raw rapl data.
func calculateDeltaPower(pre *[]uint64) []string {
	delta := make([]string, 0, 2)
	cpu, err := ghw.CPU()
	if err != nil {
		utils.Sugar.Errorf("get cpu info error: %s\n", err)
		return delta
	}
	pkgNum := len(cpu.Processors)
	cur, err := getRAPLPower(pkgNum)
	if err != nil {
		utils.Sugar.Errorf("get RAPL power error: %s\n", err)
		return delta
	}
	for i := 0; i < 2; i++ {
		orginCur := cur[i]
		if (*pre)[i] > cur[i] {
			if i == 0 {
				cur[i] += PkgMax
			} else {
				cur[i] += MemMax
			}
		}
		delta = append(delta, strconv.FormatUint(cur[i]-(*pre)[i], 10))
		(*pre)[i] = orginCur
	}

	return delta
}

// getRAPLPower reads rapl value including package and memory
func getRAPLPower(pkgNum int) ([]uint64, error) {
	res := make([]uint64, 0, pkgNum)

	for j := 0; j < 2; j++ {
		cur := uint64(0)
		path := RAPLPkgPath
		if j == 1 {
			path = RAPLMemPath
		}
		for i := 0; i < pkgNum; i++ {
			b, err := os.ReadFile(fmt.Sprintf(path, i))
			if err != nil {
				utils.Sugar.Errorf("read file error: %s\n", err)
				return nil, err
			}
			tmp, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
			if err != nil {
				utils.Sugar.Errorf("convert string to uint error: %s\n", err)
				return nil, err
			}
			cur += tmp
		}
		res = append(res, cur)
	}

	return res, nil
}
