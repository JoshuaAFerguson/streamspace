package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	streamv1alpha1 "github.com/streamspace/streamspace/api/v1alpha1"
)

var _ = Describe("Template Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a valid Template", func() {
		It("Should set status to Ready", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Valid Template",
					Description: "A valid template for testing",
					BaseImage:   "lscr.io/linuxserver/webtop:latest",
					Category:    "Desktop",
					Icon:        "https://example.com/icon.png",
					Ports: []corev1.ContainerPort{
						{
							Name:          "vnc",
							ContainerPort: 3000,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VNC: streamv1alpha1.VNCConfig{
						Enabled:  true,
						Port:     3000,
						Protocol: "websocket",
					},
					Tags: []string{"test", "desktop"},
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Verify template status becomes Ready
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "valid-template",
					Namespace: "default",
				}, createdTemplate)
				if err != nil {
					return false
				}
				return createdTemplate.Status.Valid
			}, timeout, interval).Should(Equal(true))

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})

	Context("When creating a Template without baseImage", func() {
		It("Should set status to Invalid", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Invalid Template",
					// Missing BaseImage
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Verify template status becomes Invalid
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "invalid-template",
					Namespace: "default",
				}, createdTemplate)
				if err != nil {
					return true
				}
				return createdTemplate.Status.Valid
			}, timeout, interval).Should(Equal(false))

			// Verify error message contains useful information
			Expect(createdTemplate.Status.Message).To(ContainSubstring("baseImage"))

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})

	Context("When creating a Template with VNC configuration", func() {
		It("Should validate VNC configuration", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vnc-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "VNC Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					VNC: streamv1alpha1.VNCConfig{
						Enabled:  true,
						Port:     5900,
						Protocol: "rfb",
					},
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Verify template is accepted
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "vnc-template",
					Namespace: "default",
				}, createdTemplate)
			}, timeout, interval).Should(Succeed())

			Expect(createdTemplate.Spec.VNC.Port).To(Equal(int32(5900)))
			Expect(createdTemplate.Spec.VNC.Protocol).To(Equal("rfb"))

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})

	Context("When creating a Template with WebApp configuration", func() {
		It("Should validate WebApp configuration", func() {
			Skip("WebAppConfig not yet implemented in CRD")
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "webapp-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "WebApp Template",
					BaseImage:   "nginx:latest",
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Verify template is accepted
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "webapp-template",
					Namespace: "default",
				}, createdTemplate)
			}, timeout, interval).Should(Succeed())

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})
})
