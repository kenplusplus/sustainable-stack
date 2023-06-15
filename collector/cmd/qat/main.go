package main

import (
	"sustainability.collector/pkg/qat"
	"sustainability.collector/pkg/utils"

	flag "github.com/spf13/pflag"
)

type EnergyCollectionApp struct {
	qatCollector qat.QATCollector
}

func (p *EnergyCollectionApp) Run() {

	p.qatCollector.Run()

}
func (p *EnergyCollectionApp) AddFlags() {

	flag.StringVarP(&p.qatCollector.Freq, "freq", "f", "2400000", "CPU frequency when running QAT")
	flag.StringVarP(&p.qatCollector.InputDirPath, "inputDirPath", "i", "", "Input Dir Path For QATzip")
	flag.StringVarP(&p.qatCollector.OutputDirPath, "outputDirPath", "o", "", "Output Dir Path For QATzip")
	flag.StringVarP(&p.qatCollector.ResultDirPath, "resultDirPath", "r", "", "Result Dir Path For QATzip")

	flag.Parse()
}
func main() {
	utils.InitializeLogger()

	app := &EnergyCollectionApp{}

	app.AddFlags()

	app.Run()
}
