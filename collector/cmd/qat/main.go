package main

import (
	"os"
	"os/signal"
	"syscall"

	"sustainability.collector/pkg/qat"
	"sustainability.collector/pkg/utils"

	flag "github.com/spf13/pflag"
)

type EnergyCollectionApp struct {
	qatCollector qat.QATCollector
}

func (p *EnergyCollectionApp) Run() {
	quit := make(chan struct{})
	defer close(quit)

	done := make(chan bool)

	sigs := make(chan os.Signal, 1)
	defer close(sigs)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go p.qatCollector.Run(done, quit)

	go func() {
		s := <-sigs
		quit <- struct{}{}
		utils.Sugar.Infof("Receive signal %s, exit\n", s)
	}()

	<-done
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
