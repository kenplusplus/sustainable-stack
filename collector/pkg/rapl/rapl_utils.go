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
func calculateDeltaEnergy(pre *[]uint64) []string {
	delta := make([]string, 0, 2)

	cur, err := getRAPLEnergy(pkgNum)
	if err != nil {
		utils.Sugar.Errorf("get RAPL energy error: %s\n", err)
		return delta
	}
	for i := 0; i < EnergyType; i++ {
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

// getRAPLEnergy reads rapl value including package and memory
func getRAPLEnergy(pkgNum int) ([]uint64, error) {
	res := make([]uint64, 0, pkgNum)

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

// calculateEnergy calculate the delta between two RAPL values.
func calculateEnergy(pre []uint64, cur []uint64) []uint64 {
	energy := make([]uint64, 0, EnergyType)
	for i := 0; i < EnergyType; i++ {
		var tmp uint64
		if pre[i] > cur[i] {
			if i == 0 {
				tmp = (cur[i] + PkgMax - pre[i])
			} else {
				tmp = (cur[i] + MemMax - pre[i])
			}
		} else {
			tmp = cur[i] - pre[i]
		}
		energy = append(energy, tmp)
	}
	return energy
}
