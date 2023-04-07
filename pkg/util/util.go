package util

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sgpt/pkg/kubernetes"
)

func GetParent(client *kubernetes.Client, meta metav1.ObjectMeta) (string, bool) {
	if meta.OwnerReferences != nil {
		for _, owner := range meta.OwnerReferences {
			switch owner.Kind {
			case "ReplicaSet":
				rs, err := client.GetClient().AppsV1().ReplicaSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if rs.OwnerReferences != nil {
					return GetParent(client, rs.ObjectMeta)
				}
				return "ReplicaSet/" + rs.Name, false

			case "Deployment":
				dep, err := client.GetClient().AppsV1().Deployments(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return GetParent(client, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, false

			case "StatefulSet":
				sts, err := client.GetClient().AppsV1().StatefulSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return GetParent(client, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, false

			case "DaemonSet":
				ds, err := client.GetClient().AppsV1().DaemonSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, false

			case "Ingress":
				ds, err := client.GetClient().NetworkingV1().Ingresses(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "Ingress/" + ds.Name, false
			}
		}
	}
	return meta.Name, false
}

func RemoveDuplicates(slice []string) ([]string, []string) {
	set := make(map[string]bool)
	duplicates := []string{}
	for _, val := range slice {
		if _, ok := set[val]; !ok {
			set[val] = true
		} else {
			duplicates = append(duplicates, val)
		}
	}
	uniqueSlice := make([]string, 0, len(set))
	for val := range set {
		uniqueSlice = append(uniqueSlice, val)
	}
	return uniqueSlice, duplicates
}

func SliceDiff(source, dest []string) []string {
	mb := make(map[string]struct{}, len(dest))
	for _, x := range dest {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range source {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
