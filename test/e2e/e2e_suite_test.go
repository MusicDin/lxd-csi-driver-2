package e2e

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "E2e Suite")
}

func createClient() *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", "/home/dinmusic/lxd-csi-driver-2/.kube/config")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	client, err := kubernetes.NewForConfig(config)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return client
}

var _ = ginkgo.Describe("Pod with block and FS volumes", func() {
	var client *kubernetes.Clientset
	var namespace = "default"

	ginkgo.BeforeEach(func() {
		client = createClient()
	})

	ginkgo.It("Create a pod with block and FS volumes", func() {
		ctx := context.TODO()

		sc := NewStorageClass("test-sc", "zfs-pool")
		sc.Create(ctx, client)

		pvcFS := NewPersistentVolumeClaim("test-pvc-fs", namespace, "1Gi").
			WithVolumeMode(v1.PersistentVolumeFilesystem).
			WithStorageClassName(sc.Name)
		pvcFS.Create(ctx, client)

		pvcBlock := NewPersistentVolumeClaim("test-pvc-block", namespace, "1Gi").
			WithVolumeMode(v1.PersistentVolumeBlock).
			WithStorageClassName(sc.Name)
		pvcBlock.Create(ctx, client)

		pod := NewPod("test-pod", namespace, "k8s.gcr.io/pause:3.9").
			WithPVC(*pvcFS, "/mnt/test").
			WithPVC(*pvcBlock, "/dev/vda42")
		pod.Create(ctx, client)
		pod.WaitRunning(ctx, client, 30*time.Second)

		// Cleanup.
		pod.Delete(ctx, client)
		pvcFS.Delete(ctx, client)
		pvcBlock.Delete(ctx, client)
		sc.Delete(ctx, client)
	})
})
