package analyzer

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sgpt/pkg/ai"
	"k8sgpt/pkg/kubernetes"
	"k8sgpt/pkg/util"
)

type HpaAnalyzer struct{}

func (HpaAnalyzer) RunAnalysis(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
	analysisResults *[]Analysis) error {

	list, err := client.GetClient().AutoscalingV1().HorizontalPodAutoscalers(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, hpa := range list.Items {
		var failures []string

		// check ScaleTargetRef exist
		scaleTargetRef := hpa.Spec.ScaleTargetRef
		scaleTargetRefNotFound := false

		switch scaleTargetRef.Kind {
		case "Deployment":
			_, err := client.GetClient().AppsV1().Deployments(config.Namespace).Get(ctx, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "ReplicationController":
			_, err := client.GetClient().CoreV1().ReplicationControllers(config.Namespace).Get(ctx, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "ReplicaSet":
			_, err := client.GetClient().AppsV1().ReplicaSets(config.Namespace).Get(ctx, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "StatefulSet":
			_, err := client.GetClient().AppsV1().StatefulSets(config.Namespace).Get(ctx, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		default:
			failures = append(failures, fmt.Sprintf("%s HorizontalPodAutoscaler uses %s as ScaleTargetRef which does not possible option.", hpa.Name, scaleTargetRef.Kind))
		}

		if scaleTargetRefNotFound {
			failures = append(failures, fmt.Sprintf("%s HorizontalPodAutoscaler uses %s/%s as ScaleTargetRef which does not exist.", hpa.Name, scaleTargetRef.Kind, scaleTargetRef.Name))
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)] = PreAnalysis{
				HorizontalPodAutoscalers: hpa,
				FailureDetails:           failures,
			}
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = Analysis{
			Kind:  "HorizontalPodAutoscaler",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(client, value.HorizontalPodAutoscalers.ObjectMeta)
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}
