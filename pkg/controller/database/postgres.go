package database

import (
	"context"

	databasesv1alpha1 "github.com/schemahero/schemahero/pkg/apis/databases/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileDatabase) ensurePostgresWatch(instance *databasesv1alpha1.Database) error {
	imageName := "localhost:32000/schemahero/schemahero"

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-watch",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "watch"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "watch"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "schemahero",
							Image:           imageName,
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"watch",
								"--driver",
								"postgres",
								"--uri",
								instance.Connection.Postgres.URI.Value,
								"--namespace",
								instance.Namespace,
								"--instance",
								instance.Name,
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating watch deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		return err
	} else if err != nil {
		return err
	}

	// TODO diff and update!

	return nil
}