package e2e

import (
	"context"
	"maps"
	"time"

	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/canonical/lxd-csi-driver/internal/driver"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// Pod represents a Kubernetes Pod.
type Pod struct {
	v1.Pod
}

// NewPod creates a new Pod definition with the given name, namespace, and image.
func NewPod(name string, namespace string, image string) *Pod {
	manifest := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: image,
				},
			},
		},
	}

	return &Pod{manifest}
}

// PrettyName returns the string consisting of Pod's name and namespace.
func (p *Pod) PrettyName() string {
	return PrettyName(p.Namespace, p.GenerateName, p.Name)
}

// WithPVC adds a PersistentVolumeClaim to the Pod's volumes.
// The path is the mount path inside the container for filesystem volumes
// and device path inside the container for block volumes.
func (p *Pod) WithPVC(pvc PersistentVolumeClaim, path string) *Pod {
	p.Spec.Volumes = append(p.Spec.Volumes, v1.Volume{
		Name: pvc.Name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc.Name,
			},
		},
	})

	if len(p.Spec.Containers) > 0 {
		if pvc.Spec.VolumeMode == nil || *pvc.Spec.VolumeMode == v1.PersistentVolumeFilesystem {
			// For filesystem volumes, we use the mount path.
			p.Spec.Containers[0].VolumeMounts = append(p.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
				Name:      pvc.Name,
				MountPath: path,
			})
		} else {
			// For block volumes, we use the device path.
			p.Spec.Containers[0].VolumeDevices = append(p.Spec.Containers[0].VolumeDevices, v1.VolumeDevice{
				Name:       pvc.Name,
				DevicePath: path,
			})
		}
	}

	return p
}

// WithFSVolume adds a PersistentVolumeClaim to the Pod's volumes.
func (p *Pod) WithFSVolume(persistentVolumeClaimName string, mountPath string) *Pod {
	p.Spec.Volumes = append(p.Spec.Volumes, v1.Volume{
		Name: persistentVolumeClaimName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: persistentVolumeClaimName,
			},
		},
	})

	if len(p.Spec.Containers) > 0 {
		p.Spec.Containers[0].VolumeMounts = append(p.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      persistentVolumeClaimName,
			MountPath: mountPath,
		})
	}

	return p
}

// Create creates the Pod in the Kubernetes cluster.
func (p *Pod) Create(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Create Pod " + p.PrettyName())
	newPod, err := client.CoreV1().Pods(p.Namespace).Create(ctx, &p.Pod, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	// Store generated name for future reference.
	if p.Name == "" {
		p.Name = newPod.Name
	}
}

// Delete deletes the Pod from the Kubernetes cluster.
func (p *Pod) Delete(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Delete Pod " + p.PrettyName())
	err := client.CoreV1().Pods(p.Namespace).Delete(ctx, p.Name, metav1.DeleteOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

// WaitRunning waits until the Pod is in the Running state.
func (p *Pod) WaitRunning(ctx context.Context, client *kubernetes.Clientset, timeout time.Duration) {
	ginkgo.By("Wait for Pod " + p.PrettyName() + " to be running")
	gomega.Eventually(func() bool {
		pod, err := client.CoreV1().Pods(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
		if err != nil {
			return false
		}
		return pod.Status.Phase == v1.PodRunning
	}, timeout, time.Second).Should(gomega.BeTrue())
}

// WaitDeleted waits until the Pod is deleted from the Kubernetes cluster.
func (p *Pod) WaitDeleted(ctx context.Context, client *kubernetes.Clientset, timeout time.Duration) {
	ginkgo.By("Wait for Pod " + p.PrettyName() + " to be deleted")
	gomega.Eventually(func() bool {
		_, err := client.CoreV1().Pods(p.Namespace).Get(ctx, p.Name, metav1.GetOptions{})
		return err != nil
	}, timeout, time.Second).Should(gomega.BeTrue())
}

// StorageClass represents a Kubernetes StorageClass.
type StorageClass struct {
	storagev1.StorageClass
}

// NewStorageClass creates a new StorageClass definition with the given name
// and target LXD storage pool.
func NewStorageClass(name string, lxdStoragePool string) *StorageClass {
	defaultReclaimPolicy := v1.PersistentVolumeReclaimDelete
	defaultVolumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer

	manifest := storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
		Parameters: map[string]string{
			"storagePool": lxdStoragePool,
		},
		Provisioner:       driver.DefaultDriverName,
		VolumeBindingMode: &defaultVolumeBindingMode,
		ReclaimPolicy:     &defaultReclaimPolicy,
	}

	return &StorageClass{manifest}
}

// PrettyName returns the string consisting of StorageClass's name.
func (sc *StorageClass) PrettyName() string {
	return PrettyName(sc.Namespace, sc.GenerateName, sc.Name)
}

// WithParameters allows setting additional parameters for the StorageClass.
func (sc *StorageClass) WithParameters(params map[string]string) *StorageClass {
	if sc.Parameters == nil {
		sc.Parameters = make(map[string]string)
	}

	maps.Copy(sc.Parameters, params)
	return sc
}

// WithVolumeBindingMode sets the volume binding mode for the StorageClass.
func (sc *StorageClass) WithVolumeBindingMode(mode storagev1.VolumeBindingMode) *StorageClass {
	sc.VolumeBindingMode = &mode
	return sc
}

// WithReclaimPolicy sets the reclaim policy for the StorageClass.
func (sc *StorageClass) WithReclaimPolicy(mode v1.PersistentVolumeReclaimPolicy) *StorageClass {
	sc.ReclaimPolicy = &mode
	return sc
}

// WithDefault marks the storage class as default.
func (sc *StorageClass) WithDefault(isDefault bool) *StorageClass {
	if sc.Annotations == nil {
		sc.Annotations = make(map[string]string)
	}

	key := "storageclass.kubernetes.io/is-default-class"
	if isDefault {
		sc.Annotations[key] = "true"
	} else {
		delete(sc.Annotations, key)
	}

	return sc
}

// Create creates the StorageClass in the Kubernetes cluster.
func (sc *StorageClass) Create(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Create StorageClass " + sc.PrettyName())
	newSC, err := client.StorageV1().StorageClasses().Create(ctx, &sc.StorageClass, metav1.CreateOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	if sc.Name == "" {
		sc.Name = newSC.Name
	}
}

// Delete deletes the StorageClass from the Kubernetes cluster.
func (sc *StorageClass) Delete(ctx context.Context, client *kubernetes.Clientset) {
	ginkgo.By("Delete StorageClass " + sc.PrettyName())
	err := client.StorageV1().StorageClasses().Delete(ctx, sc.Name, metav1.DeleteOptions{})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}

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
