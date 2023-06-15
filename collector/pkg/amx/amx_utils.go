package amx

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"sustainability.collector/pkg/utils"

	"github.com/google/uuid"
)

// execPerfCommand executes the perf stat command with provided parameters
func execPerfCommand(events []string, pids []int, duration int) ([]uint64, error) {
	outputFile := fmt.Sprintf("perf_%s.txt", uuid.New().String())

	cmd := exec.Command("../pkg/amx/perf_stat.sh", "-e", concatenateEvent(events),
		"-t", strconv.Itoa(duration), "-p", concatenatePid(pids),
		"-o", outputFile)
	utils.Sugar.Infof("pef stat command: %s\n", cmd.String())
	defer func() {
		err := os.Remove(outputFile)
		if err != nil {
			utils.Sugar.Errorln(err)
		}
	}()

	err := cmd.Run()
	if err != nil {
		utils.Sugar.Errorf("exec perf stat error: %s\n", err)
		return nil, err
	}

	return processRawData(events, outputFile)
}

func processRawData(events []string, filePath string) ([]uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		utils.Sugar.Errorf("open perf result file error: %s\n", err)
		return nil, err
	}
	defer file.Close()

	res := make([]uint64, len(events))
	scanner := bufio.NewScanner(file)
	for i, event := range events {
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), event) {
				tmp := strings.Trim(scanner.Text(), " ")
				splited := strings.Split(tmp, " ")
				val, err := parseCountValue(splited[0])
				if err != nil {
					utils.Sugar.Errorf("parse string to uint error: %s\n", err)
					return nil, err
				}
				res[i] = val
				break
			}
		}
	}

	return res, err
}

func concatenateEvent(events []string) string {
	var res string
	for _, v := range events {
		res += v + ","
	}
	return strings.TrimSuffix(res, ",")
}

func concatenatePid(pids []int) string {
	var res string
	for _, v := range pids {
		res += strconv.Itoa(v) + ","
	}
	return strings.TrimSuffix(res, ",")
}

func parseCountValue(count string) (uint64, error) {
	s := strings.Split(count, ",")
	res := uint64(0)
	for _, v := range s {
		res *= 1000
		tmp, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return res, err
		}
		res += tmp
	}
	return res, nil
}

// setCPUFreq sets the cpu frequency
func setCpuFreq(freq int) error {
	cmd := exec.Command("cpupower", "frequency-set", "--max", strconv.Itoa(freq), "--min", strconv.Itoa(freq))
	err := cmd.Run()
	if err != nil {
		utils.Sugar.Errorf("set cpu frequency error: %s\n", err)
		return err
	}
	return nil
}
