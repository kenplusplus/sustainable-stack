package rapl

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sustainability.collector/pkg/utils"
)

// calculateDeltaEnergy computes the rapl delta value compared with previous value,
// and normalizes the raw rapl data.
func calculateDeltaEnergy(pre *[]uint64) [EnergyType]string {
	var delta [EnergyType]string
	cur, err := getRAPLEnergy(pkgNum)
	if err != nil {
		utils.Sugar.Errorf("get RAPL energy error: %s\n", err)
		return delta
	}
	for i := 0; i < EnergyType; i++ {
		orginCur := cur[i]
		if (*pre)[i] > cur[i] {
			if i == 0 {
				cur[i] += pkgMax
			} else {
				cur[i] += memMax
			}
		}
		delta[i] = strconv.FormatUint(cur[i]-(*pre)[i], 10)
		(*pre)[i] = orginCur
	}

	return delta
}

// getRAPLEnergy reads rapl value including package and memory
func getRAPLEnergy(pkgNum int) ([EnergyType]uint64, error) {
	var res [EnergyType]uint64
	for j := 0; j < EnergyType; j++ {
		cur := uint64(0)
		path := RAPLPkgPath
		if j == 1 {
			path = RAPLMemPath
		}
		for i := 0; i < pkgNum; i++ {
			b, err := os.ReadFile(fmt.Sprintf(path, i))
			if err != nil {
				utils.Sugar.Errorf("read file error: %s\n", err)
				return res, err
			}
			tmp, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
			if err != nil {
				utils.Sugar.Errorf("convert string to uint error: %s\n", err)
				return res, err
			}
			cur += tmp
		}
		res[j] = cur
	}

	return res, nil
}

// calculateEnergy calculate the delta between two RAPL values.
func calculateEnergy(pre [EnergyType]uint64, cur [EnergyType]uint64) [EnergyType]uint64 {
	var energy [EnergyType]uint64
	for i := 0; i < EnergyType; i++ {
		var tmp uint64
		if pre[i] > cur[i] {
			if i == 0 {
				tmp = (cur[i] + pkgMax - pre[i])
			} else {
				tmp = (cur[i] + memMax - pre[i])
			}
		} else {
			tmp = cur[i] - pre[i]
		}
		energy[i] = tmp
	}

	return energy
}

// getMaxValue get max RAPL value from RAPLMaxValuesPath
func getMaxValue(path string) (uint64, error) {
	maxValue, err := os.ReadFile(fmt.Sprintf(path, 0))
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(maxValue)), 10, 64)
}
