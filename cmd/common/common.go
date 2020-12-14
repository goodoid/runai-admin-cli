package common

import (
	"os"

	"github.com/run-ai/runai-cli/pkg/client"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NumberOfRetiresForApiServer = 3
)

func ScaleRunaiOperator(client *client.Client, replicas int32) {
	deployment, err := client.GetClientset().AppsV1().Deployments("runai").Get("runai-operator", metav1.GetOptions{})
	if err != nil {
		log.Infof("Failed to get runai operator, error: %v", err)
		os.Exit(1)
	}
	for i := 0; i < NumberOfRetiresForApiServer; i++ {
		deployment.Spec.Replicas = &replicas
		deployment, err = client.GetClientset().AppsV1().Deployments("runai").Update(deployment)
		if err != nil {
			log.Debugf("Failed to update runai operator, attempt: %v, error: %v", i, err)
			continue
		}
		break
	}

	if err != nil {
		log.Infof("Failed to update runai operator, error: %v", err)
		os.Exit(1)
	}
	log.Infof("Scaled RunAI Operator to: %v", replicas)
}
