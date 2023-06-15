package rest

import (
	"context"

	"sustainability.collector/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeInfo struct {
	KubeConfigPath string
	Namespace      string
	Deployment     string
	ScaleNum       int
}

// ScaleAMXPods scales AMX infer pods to the target number
func (p *KubeInfo) ScaleAMXPods() error {
	config, err := clientcmd.BuildConfigFromFlags("", p.KubeConfigPath)
	if err != nil {
		utils.Sugar.Errorf("load kubeconfig error: %s\n", config)
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		utils.Sugar.Errorf("load kubeconfig error: %s\n", err)
		return err
	}

	// get the current target scale obj
	s, err := clientset.AppsV1().Deployments(p.Namespace).
		GetScale(context.TODO(), p.Deployment, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		utils.Sugar.Errorf("Deployment %s in namespace %s is not found\n", p.Deployment, p.Namespace)
		return err
	} else if err != nil {
		utils.Sugar.Errorf("scale pods error: %s\n", err)
		return err
	}

	if p.ScaleNum <= 0 {
		p.ScaleNum = int(s.Spec.Replicas)
		utils.Sugar.Infoln("Not set scale number of AMX pods, so do not scale it")
		return nil
	}

	sc := *s
	sc.Spec.Replicas = int32(p.ScaleNum)

	// scale pods
	_, err = clientset.AppsV1().Deployments(p.Namespace).
		UpdateScale(context.TODO(), p.Deployment, &sc, metav1.UpdateOptions{})
	if err != nil {
		utils.Sugar.Errorf("scale pods error: %s\n", err)
		return err
	}

	return nil
}
