package common

import (
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NumberOfRetiresForApiServer = 3
)

func ScaleRunaiOperator(client *client.Client, replicas int32) {
	var err error
	var deployment *appsv1.Deployment
	for i := 0; i < NumberOfRetiresForApiServer; i++ {
		deployment, err = client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
		if err != nil {
			log.Infof("Failed to get Run:AI operator, error: %v", err)
			os.Exit(1)
		}
		deployment.Spec.Replicas = &replicas
		deployment, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
		if err != nil {
			log.Debugf("Failed to update Run:AI operator, attempt: %v, error: %v", i, err)
			continue
		}
		break
	}
	if err != nil {
		log.Infof("Failed to update Run:AI operator, error: %v", err)
		os.Exit(1)
	}
	log.Infof("Scaled Run:AI operator to: %v", replicas)
}
