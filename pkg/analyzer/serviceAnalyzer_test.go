package analyzer

import (
	"context"
	"testing"

	"github.com/magiconair/properties/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8sgpt/pkg/kubernetes"
)

func TestServiceAnalzyer(t *testing.T) {

	clientset := fake.NewSimpleClientset(&v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "example",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"app": "example",
				},
			}})

	serviceAnalyzer := ServiceAnalyzer{}
	var analysisResults []Analysis
	serviceAnalyzer.RunAnalysis(context.Background(),
		&AnalysisConfiguration{
			Namespace: "default",
		},
		&kubernetes.Client{
			Client: clientset,
		}, nil, &analysisResults)

	assert.Equal(t, len(analysisResults), 1)
}
