package qat

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"sustainability.collector/pkg/rapl"
	"sustainability.collector/pkg/utils"
)

// preProcess preparation work:1.set cpu frequency; 2.get idle power; 3.get input files
func preProcess(frequency string) ([]string, []uint64, error) {

	var inputFiles []string
	//set cpu freqyency
	err := setCpuFreq(frequency)
	if err != nil {
		utils.Sugar.Panicln(err)
		return nil, nil, err
	}
	//get idle power
	idlePower, err := rapl.GetIdlePower(idleTime)
	if err != nil {
		utils.Sugar.Errorf("failed to get idle power : %s\n", err)
		return nil, nil, err
	}

	//traverse the folder and obtain the files that need to be processed
	err = filepath.Walk(inputDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			inputFiles = append(inputFiles, path)
		} else {
			utils.Sugar.Errorln("input folder is empty.")
		}
		return nil
	})

	if err != nil {
		utils.Sugar.Panicf("failed to read input directory: %s\n", err)
	}

	return inputFiles, idlePower, err
}

// preCollector collect base data and qzip args from the input file.
func preCollector(inputFile string, frequency string) ([]string, []string) {

	var (
		mode       int
		args       string
		outputFile string
		qzipArgs   = make([]string, 0, 3)
		baseData   = make([]string, 0, 3)
	)

	if strings.HasSuffix(inputFile, ".gz") {
		//the input file need to decompress
		mode = DecompressMode

		//set qzip args
		args = "-d "

		//the compressed file extension format changes to the decompressed file extension
		base := filepath.Base(inputFile)
		filename := strings.TrimSuffix(base, ".gz")
		outputFile = filepath.Join(outputDirPath, filename)

	} else {
		//the input file need to compress
		mode = CompressMode

		//the decompressed file extension format changes to the compressed file extension
		base := filepath.Base(inputFile)
		outputFile = filepath.Join(outputDirPath, base)
	}

	baseData = append(baseData, inputFile, frequency, strconv.Itoa(mode))
	qzipArgs = append(qzipArgs, args, inputFile, outputFile)

	return baseData, qzipArgs
}

// structToStringSlice convert a struct to a string slice
func structToStringSlice(s interface{}) []string {
	v := reflect.ValueOf(s)
	t := v.Type()

	data := make([]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		switch field.Kind() {
		case reflect.String:
			data = append(data, field.String())
		case reflect.Uint64:
			data = append(data, strconv.FormatUint(field.Uint(), 10))
		case reflect.Float64:
			data = append(data, strconv.FormatFloat(field.Float(), 'f', 3, 64))
		default:
			// Handle other field types as needed
			data = append(data, fmt.Sprintf("%v", field.Interface()))
		}
	}

	return data
}

// setCPUFreq change the cpu frequency
func setCpuFreq(freq string) error {
	cmd := exec.Command("cpupower", "frequency-set", "--max", freq, "--min", freq)
	err := cmd.Run()
	if err != nil {
		utils.Sugar.Errorf("set cpu frequency error: %s\n", err)
		return err
	}
	return nil
}

// closeChannel check channel and close
func closeChannel(ch chan struct{}) {
	select {
	case _, ok := <-ch:
		if !ok {
			utils.Sugar.Infoln("channel is already closed")
		}
	default:
		close(ch)
	}
}
