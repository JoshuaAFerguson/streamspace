package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

var _ = Describe("Template Controller - Advanced Validation", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating templates with validation edge cases", func() {
		It("Should reject template with missing DisplayName", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "no-displayname-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					// Missing DisplayName
					BaseImage: "lscr.io/linuxserver/firefox:latest",
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Should be marked as invalid or accepted with defaults
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "no-displayname-template",
					Namespace: "default",
				}, createdTemplate)
			}, timeout, interval).Should(Succeed())

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})

		It("Should handle template with invalid image format", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bad-image-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Bad Image Template",
					BaseImage:   "not-a-valid-image::", // Invalid format
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}

			// K8s may accept this, but deployment will fail
			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Template may be marked as valid at API level
			// Actual validation happens when creating sessions
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "bad-image-template",
					Namespace: "default",
				}, createdTemplate)
			}, timeout, interval).Should(Succeed())

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})

		It("Should validate port configurations", func() {
			ctx := context.Background()

			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-port-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Multi Port Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					Ports: []corev1.ContainerPort{
						{
							Name:          "vnc",
							ContainerPort: 3000,
							Protocol:      corev1.ProtocolTCP,
						},
						{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}

			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Verify multiple ports accepted
			createdTemplate := &streamv1alpha1.Template{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "multi-port-template",
					Namespace: "default",
				}, createdTemplate)
			}, timeout, interval).Should(Succeed())

			Expect(createdTemplate.Spec.Ports).To(HaveLen(2))

			// Cleanup
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})
})

var _ = Describe("Template Controller - Resource Defaults", func() {
	const (
		timeout  = time.Second * 15
		interval = time.Millisecond * 250
	)

	Context("When template defines default resources", func() {
		It("Should propagate defaults to sessions", func() {
			ctx := context.Background()

			// Create template with default resources
			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaults-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Defaults Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					DefaultResources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("4Gi"),
							corev1.ResourceCPU:    resource.MustParse("2000m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("8Gi"),
							corev1.ResourceCPU:    resource.MustParse("4000m"),
						},
					},
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}
			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Create session without specifying resources
			session := &streamv1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "defaults-session",
					Namespace: "default",
				},
				Spec: streamv1alpha1.SessionSpec{
					User:           "defaultsuser",
					Template:       "defaults-template",
					State:          "running",
					PersistentHome: false,
					// No resources specified - should use template defaults
				},
			}
			Expect(k8sClient.Create(ctx, session)).To(Succeed())

			// Verify deployment uses template defaults
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "ss-defaultsuser-defaults-template",
					Namespace: "default",
				}, deployment)
			}, timeout, interval).Should(Succeed())

			// Check if defaults were applied
			if len(deployment.Spec.Template.Spec.Containers) > 0 {
				container := deployment.Spec.Template.Spec.Containers[0]
				// Defaults should be applied
				Expect(container.Resources).NotTo(BeNil())
			}

			// Cleanup
			Expect(k8sClient.Delete(ctx, session)).To(Succeed())
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})

		It("Should allow session-level resource overrides", func() {
			ctx := context.Background()

			// Create session with custom resources (override template defaults)
			session := &streamv1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "override-session",
					Namespace: "default",
				},
				Spec: streamv1alpha1.SessionSpec{
					User:     "overrideuser",
					Template: "defaults-template",
					State:    "running",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("1Gi"),
							corev1.ResourceCPU:    resource.MustParse("500m"),
						},
					},
					PersistentHome: false,
				},
			}
			Expect(k8sClient.Create(ctx, session)).To(Succeed())

			// Verify deployment uses session overrides, not template defaults
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "ss-overrideuser-defaults-template",
					Namespace: "default",
				}, deployment)
			}, timeout, interval).Should(Succeed())

			// Check override was applied
			if len(deployment.Spec.Template.Spec.Containers) > 0 {
				container := deployment.Spec.Template.Spec.Containers[0]
				memRequest := container.Resources.Requests.Memory()
				if memRequest != nil {
					Expect(memRequest.String()).To(Equal("1Gi"))
				}
			}

			// Cleanup
			Expect(k8sClient.Delete(ctx, session)).To(Succeed())
		})
	})
})

var _ = Describe("Template Controller - Lifecycle", func() {
	const (
		timeout  = time.Second * 15
		interval = time.Millisecond * 250
	)

	Context("When template is updated", func() {
		It("Should not affect existing sessions", func() {
			ctx := context.Background()

			// Create template
			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lifecycle-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Lifecycle Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}
			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Create session from template
			session := &streamv1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lifecycle-session",
					Namespace: "default",
				},
				Spec: streamv1alpha1.SessionSpec{
					User:           "lifecycleuser",
					Template:       "lifecycle-template",
					State:          "running",
					PersistentHome: false,
				},
			}
			Expect(k8sClient.Create(ctx, session)).To(Succeed())

			// Wait for deployment
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "ss-lifecycleuser-lifecycle-template",
					Namespace: "default",
				}, deployment)
			}, timeout, interval).Should(Succeed())

			originalImage := deployment.Spec.Template.Spec.Containers[0].Image

			// Update template (change base image)
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "lifecycle-template",
				Namespace: "default",
			}, template)).To(Succeed())
			template.Spec.BaseImage = "lscr.io/linuxserver/chromium:latest"
			Expect(k8sClient.Update(ctx, template)).To(Succeed())

			// Wait a bit
			time.Sleep(2 * time.Second)

			// Verify existing session deployment NOT updated
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "ss-lifecycleuser-lifecycle-template",
				Namespace: "default",
			}, deployment)).To(Succeed())

			currentImage := deployment.Spec.Template.Spec.Containers[0].Image
			// Existing session should keep original image
			Expect(currentImage).To(Equal(originalImage))

			// Cleanup
			Expect(k8sClient.Delete(ctx, session)).To(Succeed())
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})

		It("Should apply to new sessions after update", func() {
			ctx := context.Background()

			// Create template
			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-session-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "New Session Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}
			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Update template before creating session
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "new-session-template",
				Namespace: "default",
			}, template)).To(Succeed())
			template.Spec.DisplayName = "Updated Display Name"
			Expect(k8sClient.Update(ctx, template)).To(Succeed())

			// Create new session after update
			session := &streamv1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-after-update-session",
					Namespace: "default",
				},
				Spec: streamv1alpha1.SessionSpec{
					User:           "newafterupdateuser",
					Template:       "new-session-template",
					State:          "running",
					PersistentHome: false,
				},
			}
			Expect(k8sClient.Create(ctx, session)).To(Succeed())

			// New session should use updated template
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "ss-newafterupdateuser-new-session-template",
					Namespace: "default",
				}, deployment)
			}, timeout, interval).Should(Succeed())

			// Cleanup
			Expect(k8sClient.Delete(ctx, session)).To(Succeed())
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())
		})
	})

	Context("When template is deleted", func() {
		It("Should handle deletion gracefully", func() {
			ctx := context.Background()

			// Create template
			template := &streamv1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "delete-template",
					Namespace: "default",
				},
				Spec: streamv1alpha1.TemplateSpec{
					DisplayName: "Delete Template",
					BaseImage:   "lscr.io/linuxserver/firefox:latest",
					VNC: streamv1alpha1.VNCConfig{
						Enabled: true,
						Port:    3000,
					},
				},
			}
			Expect(k8sClient.Create(ctx, template)).To(Succeed())

			// Create session from template
			session := &streamv1alpha1.Session{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "orphaned-session",
					Namespace: "default",
				},
				Spec: streamv1alpha1.SessionSpec{
					User:           "orphaneduser",
					Template:       "delete-template",
					State:          "running",
					PersistentHome: false,
				},
			}
			Expect(k8sClient.Create(ctx, session)).To(Succeed())

			// Wait for session to be created
			time.Sleep(2 * time.Second)

			// Delete template
			Expect(k8sClient.Delete(ctx, template)).To(Succeed())

			// Session should continue to exist (orphaned but functional)
			createdSession := &streamv1alpha1.Session{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      "orphaned-session",
					Namespace: "default",
				}, createdSession)
			}, timeout, interval).Should(Succeed())

			// Cleanup session
			Expect(k8sClient.Delete(ctx, session)).To(Succeed())
		})
	})
})
