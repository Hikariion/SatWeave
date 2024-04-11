package watcher

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"satweave/utils/logger"
)

// CreateJob 创建Job
func CreateJob(jobName, namespace, imageName string, args []string) error {
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  jobName,
							Image: imageName,
							Args:  args,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	jobsClient := clientset.BatchV1().Jobs(namespace)
	_, err = jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logger.Infof("Job %s created successfully", jobName)
	return nil
}

// GetAllNamespaces
// for test
func GetAllNamespaces() ([]string, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}

	return nsNames, nil
}

// UploadFile 上传文件
