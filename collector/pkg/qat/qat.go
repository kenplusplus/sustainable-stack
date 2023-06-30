package qat

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"

	"sustainability.collector/pkg/rapl"
	"sustainability.collector/pkg/telemetry"
	"sustainability.collector/pkg/utils"

	"sync"
	"time"
)

const (
	idleTime = 1 * time.Minute
)

var (
	inputDirPath  string
	outputDirPath string
	resultDirPath string
)

type QATCollector struct {
	Freq          string
	InputDirPath  string
	OutputDirPath string
	ResultDirPath string
}

func (q *QATCollector) Run(done chan bool, quit chan struct{}) {

	inputDirPath = q.InputDirPath
	outputDirPath = q.OutputDirPath
	resultDirPath = q.ResultDirPath

	//preprocessing step
	inputFiles, idlePower, err := preProcess(q.Freq)
	if err != nil {
		utils.Sugar.Errorln(err)
		return
	}

	//prepare csvfile and write column headers
	err = writeColumnHeaders()
	if err != nil {
		utils.Sugar.Errorln(err)
		return
	}

	baseData := []string{q.Freq}

	for _, inputFile := range inputFiles {

		qzipArgs := preCollector(inputFile)

		err = collectData(baseData, qzipArgs, idlePower, quit, done)
		if err != nil {
			utils.Sugar.Errorln(err)
			continue
		}
	}

	utils.Sugar.Infoln("Collection complete.")
	done <- true
}

// collectData collect telemetry value & rapl value during compression/decompression
// and write into csv file
func collectData(baseData []string, qzipArgs []string, idlePower []uint64, quit chan struct{}, done chan bool) error {
	var (
		wg      sync.WaitGroup
		resTele *telemetry.ResTelemetry
		resRapl []string
	)

	//open result csv file
	resultPath := filepath.Join(resultDirPath, "result.csv")
	resultFile, err := os.OpenFile(resultPath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		utils.Sugar.Panicf("failed to open result file: %s", err)
	}
	defer resultFile.Close()

	writer := csv.NewWriter(resultFile)

	stopChan := make(chan struct{})

	defer closeChannel(stopChan)

	wg.Add(1)

	//collect telemetry value
	go func() {
		defer wg.Done()
		resTele, err = telemetry.ReadTelemetry(stopChan)
		if err != nil {
			utils.Sugar.Errorln(err)
		}
	}()

	select {

	case <-quit:
		closeChannel(stopChan)
		utils.Sugar.Infoln("receive a signal, stop collecting")

		//wait for stop telemetry
		time.Sleep(1 * time.Second)
		done <- false
	default:
		//collect rapl value
		resRapl, err = raplCollector(qzipArgs, idlePower)
		if err != nil {
			utils.Sugar.Errorf("failed to collector rapl value:%s \n", err)
			return err
		}
	}

	//wait for telemetry
	time.Sleep(5 * time.Second)

	// stop telemetry
	closeChannel(stopChan)
	wg.Wait()

	resData := append(baseData, append(structToStringSlice(*resTele), resRapl...)...)

	// store the result data to csv file
	err = writer.Write(resData)
	if err != nil {
		utils.Sugar.Errorf("failed to write result data :%s \n", err)
		return err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		utils.Sugar.Errorf("failed to flush data :%s \n", err)
		return err
	}

	return nil
}

// raplCollector collect rapl value during compression/decompression
// and normalize the raw data
func raplCollector(qzipArgs []string, idlePower []uint64) ([]string, error) {
	q := &rapl.RAPLEnergy{}

	//open qzip log file
	qzipLogPath := filepath.Join(resultDirPath, "qzip.log")
	qzipLogFile, err := os.OpenFile(qzipLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		utils.Sugar.Panicf("failed to open qziplog file:%s \n", err)
		return nil, err
	}
	defer qzipLogFile.Close()

	// read rapl
	preEnergy, err := q.ReadCurrentRapl()
	if err != nil {
		utils.Sugar.Errorf("failed to read pre rapl:%s \n", err)
		return nil, err
	}

	//  start the timer
	start := time.Now()

	commandText := "qzip -O gzip -k " + qzipArgs[0] + qzipArgs[1] + " -o " + qzipArgs[2]

	cmd := exec.Command("bash", "-c", commandText)
	utils.Sugar.Infoln("start compress/decompress file")

	out, err := cmd.CombinedOutput()
	if err != nil {
		utils.Sugar.Errorln(err)
		return nil, err
	}

	// stop the timer
	end := time.Since(start).Seconds()

	// read current rapl
	curEnergy, err := q.ReadCurrentRapl()
	if err != nil {
		utils.Sugar.Errorf("failed to read current rapl:%s \n", err)
		return nil, err
	}

	//store qzip log
	_, err = qzipLogFile.Write(out)
	if err != nil {
		utils.Sugar.Errorf("write qzip log failed:%s \n", err)
	}

	//calculate energy
	resEnergy := q.CalculateDynEnergy(idlePower, preEnergy, curEnergy, end)

	return resEnergy[:], nil
}
