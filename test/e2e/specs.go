package e2e

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// PersistentVolumeClaim represents a Kubernetes PersistentVolumeClaim.
type PersistentVolumeClaim struct {
	v1.PersistentVolumeClaim
}

// NewPersistentVolumeClaim creates a new PersistentVolumeClaim with the given name,
// namespace, and size. The size can be specified in bytes or in binary SI format.
// It default to ReadWriteOnce access mode.
func NewPersistentVolumeClaim(name string, namespace string, sizeBytes string) *PersistentVolumeClaim {
	manifest := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(sizeBytes),
				},
			},
		},
	}

	return &PersistentVolumeClaim{manifest}
}

// PrettyName returns the string consisting of PersistentVolumeClaim's name and namespace.
func (pvc *PersistentVolumeClaim) PrettyName() string {
	return PrettyName(pvc.Namespace, pvc.GenerateName, pvc.Name)
}

// WithVolumeMode sets the volume mode for the PersistentVolumeClaim.
// It can be either Filesystem or Block.
func (pvc *PersistentVolumeClaim) WithVolumeMode(mode v1.PersistentVolumeMode) *PersistentVolumeClaim {
	pvc.Spec.VolumeMode = &mode
	return pvc
}

// WithAccessModes sets the access modes for the PersistentVolumeClaim.
func (pvc *PersistentVolumeClaim) WithAccessModes(accessModes ...v1.PersistentVolumeAccessMode) *PersistentVolumeClaim {
	pvc.Spec.AccessModes = accessModes
	return pvc
}

// WithStorageClassName sets the storage class name for the PersistentVolumeClaim.
func (pvc *PersistentVolumeClaim) WithStorageClassName(storageClassName string) *PersistentVolumeClaim {
	pvc.Spec.StorageClassName = &storageClassName
	return pvc
}

// Create creates the PersistentVolumeClaim in the Kubernetes cluster.
func (pvc *PersistentVolumeClaim) Create(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Create PersistentVolumeClaim " + pvc.PrettyName())
	newPVC, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, &pvc.PersistentVolumeClaim, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// Store generated name for future reference.
	if pvc.Name == "" {
		pvc.Name = newPVC.Name
	}
}

// Delete deletes the PersistentVolumeClaim from the Kubernetes cluster.
func (pvc *PersistentVolumeClaim) Delete(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Delete PersistentVolumeClaim " + pvc.PrettyName())
	err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Delete(ctx, pvc.Name, metav1.DeleteOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// WaitBound waits until the PersistentVolumeClaim is bound to a PersistentVolume.
func (pvc *PersistentVolumeClaim) WaitBound(ctx context.Context, client *kubernetes.Clientset, timeout time.Duration) {
	ginkgo.By("Wait for PersistentVolumeClaim " + pvc.PrettyName() + " to be bound")
	gomega.Eventually(func() bool {
		pvc, err := client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(ctx, pvc.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return pvc.Status.Phase == v1.ClaimBound
	}, timeout, time.Second).Should(gomega.BeTrue())
}
