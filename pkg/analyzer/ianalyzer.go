package analyzer

import (
	"context"

	"k8sgpt/pkg/ai"
	"k8sgpt/pkg/kubernetes"
)

type IAnalyzer interface {
	RunAnalysis(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
		analysisResults *[]Analysis) error
}
