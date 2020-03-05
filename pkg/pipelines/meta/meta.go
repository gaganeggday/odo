package meta

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TypeMeta creates v1.TypeMeta
func TypeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// ObjectMeta creates v1.ObjectMeta
func ObjectMeta(n types.NamespacedName) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
}

// NamespacedName creates types.NamespacedName
func NamespacedName(ns, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
}
