package qat

import (
	"encoding/csv"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"sustainability.collector/pkg/rapl"
	"sustainability.collector/pkg/telemetry"
	"sustainability.collector/pkg/utils"

	"sync"
	"time"
)

const (
	CompressMode = iota
	DecompressMode
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

func (q *QATCollector) Run() {

	inputDirPath = q.InputDirPath
	outputDirPath = q.OutputDirPath
	resultDirPath = q.ResultDirPath

	//preprocessing step
	inputFiles, idlePower, err := preProcess(q.Freq)
	if err != nil {
		utils.Sugar.Errorln(err)
		return
	}

	for _, inputFile := range inputFiles {

		baseData, qzipArgs := preCollector(inputFile, q.Freq)

		err, sigs := collectData(baseData, qzipArgs, idlePower)
		if err != nil {
			utils.Sugar.Errorln(err)
			continue
		} else if sigs != nil {
			break
		}
	}

}

// collectData collect telemetry value & rapl value during compression/decompression
// and write into csv file
func collectData(baseData []string, qzipArgs []string, idlePower []uint64) (error, os.Signal) {
	var (
		wg      sync.WaitGroup
		resTele *telemetry.ResTelemetry
		s       os.Signal
	)

	//open result csv file
	resultPath := filepath.Join(resultDirPath, "result.csv")
	resultFile, err := os.OpenFile(resultPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		utils.Sugar.Panicf("failed to open result file: %s", err)
	}
	defer resultFile.Close()

	writer := csv.NewWriter(resultFile)

	stopChan := make(chan struct{})
	defer closeChannel(stopChan)

	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)

	//collect telemetry value
	go func() {
		defer wg.Done()
		resTele, err = telemetry.ReadTelemetry(stopChan)
		if err != nil {
			utils.Sugar.Errorln(err)
		}
	}()

	go func() {
		s = <-sigChan
		utils.Sugar.Infof("receive signal %s, stop collect\n", s)
	}()

	//collect rapl value
	resRapl, err := raplCollector(qzipArgs, idlePower)
	if err != nil {
		utils.Sugar.Errorf("failed to collector rapl value:%s \n", err)
		return err, nil
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
		return err, nil
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		utils.Sugar.Errorf("failed to flush data :%s \n", err)
		return err, nil
	}

	return nil, s
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

	result := append([]string{strconv.FormatFloat(end, 'f', 3, 64)}, resEnergy...)

	return result, nil
}
