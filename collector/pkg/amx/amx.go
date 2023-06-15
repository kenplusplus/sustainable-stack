package amx

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"sustainability.collector/pkg/utils"
)

type AMXCollector struct {
	Events      []string
	Pids        []int
	Freq        int
	InferPodNum int
}

func (a *AMXCollector) Run(quit chan struct{}) {
	// open csv file
	fileName := fmt.Sprintf("amx_event_%s.csv", time.Now().Format("20060102150405"))
	f, err := os.Create(fileName)
	if err != nil {
		utils.Sugar.Panicf("create csv file error: %s", err)
	}
	defer f.Close()

	// read amx_busy performance counter value
	HWCount := make(chan []string)
	defer close(HWCount)
	go a.getHWCounterHelper(HWCount, quit, 60*time.Second)

	copyed := make([]string, len(a.Events))
	copy(copyed, a.Events)
	copyed = append(copyed, "cpu_freq", "inferpod_num")

	err = setCpuFreq(a.Freq)
	if err != nil {
		utils.Sugar.Panicln(err)
	}
	// store data to csv file
	go func() {
		w := csv.NewWriter(f)
		if err = w.Write(copyed); err != nil {
			utils.Sugar.Errorf("error writing column header to csv: %s\n", err)
		}
		for v := range HWCount {
			if err := w.Write(v); err != nil {
				utils.Sugar.Errorf("error writing record to csv: %s\n", err)
			}
			w.Flush()
		}
	}()

	<-quit
}

// getHWCounterHepler that periodically collects the value of the target events
func (a *AMXCollector) getHWCounterHelper(HWCount chan []string, quit chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			counts, err := execPerfCommand(a.Events, a.Pids, int(interval/time.Second))
			if err != nil {
				utils.Sugar.Panicln(err)
				continue
			}
			res := make([]string, 0, len(counts))
			for _, v := range counts {
				res = append(res, strconv.FormatUint(v, 10))
			}
			res = append(append(res, strconv.Itoa(a.Freq)), strconv.Itoa(a.InferPodNum))
			HWCount <- res
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
