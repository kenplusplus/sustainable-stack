package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sustainability.collector/pkg/amx"
	"sustainability.collector/pkg/rapl"
	"sustainability.collector/pkg/rest"
	"sustainability.collector/pkg/utils"

	flag "github.com/spf13/pflag"
	"k8s.io/client-go/util/homedir"
)

type PowerCollectionApp struct {
	amxCollector amx.AMXCollector
	scaleInfo    rest.KubeInfo
}

func (p *PowerCollectionApp) Run() {
	utils.Sugar.Infow("Scale amx infer pods",
		"namespace", p.scaleInfo.Namespace,
		"deployment", p.scaleInfo.Deployment,
		"replicas", p.scaleInfo.ScaleNum)

	err := p.scaleInfo.ScaleAMXPods()
	if err != nil {
		utils.Sugar.Errorln(err)
		return
	}

	// waits infer pods ready
	time.Sleep(2 * time.Minute)

	quit := make(chan struct{})
	defer close(quit)
	sigs := make(chan os.Signal, 1)
	defer close(sigs)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	raplObj := &rapl.RAPLEnergy{}

	go raplObj.Run(quit)

	p.amxCollector.InferPodNum = p.scaleInfo.ScaleNum

	go p.amxCollector.Run(quit)

	s := <-sigs
	utils.Sugar.Infof("Receive signal %s, exit\n", s)

}

func (p *PowerCollectionApp) AddFlags() {
	// parameters for collecting HW counters
	flag.StringSliceVarP(&p.amxCollector.Events, "events", "e", []string{}, "PMU events that need to be monitored")
	flag.IntSliceVarP(&p.amxCollector.Pids, "pids", "p", []int{}, "PIDs that need to be monitored")
	flag.IntVarP(&p.amxCollector.Freq, "freq", "f", 2400000, "CPU frequency when running AMX")

	// kubernetes related parameters
	flag.StringVarP(&p.scaleInfo.Namespace, "namespace", "n", "default", "Namespace where the scaling deployment in")
	flag.StringVarP(&p.scaleInfo.Deployment, "deployName", "d", "", "The deployment that needs to be scaled")
	flag.IntVar(&p.scaleInfo.ScaleNum, "pods", -1, "Number of AMX running pods")

	// load kubeconfig
	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&p.scaleInfo.KubeConfigPath, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&p.scaleInfo.KubeConfigPath, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()
}

func main() {
	utils.InitializeLogger()

	app := &PowerCollectionApp{}

	app.AddFlags()

	app.Run()
}
