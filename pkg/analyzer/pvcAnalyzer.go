package analyzer

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sgpt/pkg/ai"
	"k8sgpt/pkg/kubernetes"
	"k8sgpt/pkg/util"
)

type PvcAnalyzer struct{}

func (PvcAnalyzer) RunAnalysis(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().PersistentVolumeClaims(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, pvc := range list.Items {
		var failures []string

		// Check for empty rs
		if pvc.Status.Phase == "Pending" {

			// parse the event log and append details
			evt, err := FetchLatestEvent(ctx, client, pvc.Namespace, pvc.Name)
			if err != nil || evt == nil {
				continue
			}
			if evt.Reason == "ProvisioningFailed" && evt.Message != "" {
				failures = append(failures, evt.Message)
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)] = PreAnalysis{
				PersistentVolumeClaim: pvc,
				FailureDetails:        failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = Analysis{
			Kind:  "PersistentVolumeClaim",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(client, value.PersistentVolumeClaim.ObjectMeta)
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}
