/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	"github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"os"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	instancev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/api/v1"
	templatev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1"
)

// LabInstanceReconciler reconciles a LabInstance object
type LabInstanceReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	EventsRecorder record.EventRecorder
}

// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events/status,verbs=get

func (r *LabInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("labinstance", req.NamespacedName)

	// get labInstance
	var labInstance instancev1.LabInstance
	if err := r.Get(ctx, req.NamespacedName, &labInstance); err != nil {
		// reconcile was triggered by a delete request
		log.Info("LabInstance " + req.Name + " deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// The metadata.generation value is incremented for all changes, except for changes to .metadata or .status
	// if metadata.generation is not incremented there's no need to reconcile
	if labInstance.Status.ObservedGeneration == labInstance.ObjectMeta.Generation {
		return ctrl.Result{}, nil
	}

	// check if labTemplate exists
	templateName := types.NamespacedName{
		Namespace: labInstance.Spec.LabTemplateNamespace,
		Name:      labInstance.Spec.LabTemplateName,
	}
	var labTemplate templatev1.LabTemplate
	if err := r.Get(ctx, templateName, &labTemplate); err != nil {
		// no LabTemplate related exists
		log.Info("LabTemplate " + templateName.Name + " doesn't exist. Deleting LabInstance " + labInstance.Name)
		r.EventsRecorder.Event(&labInstance, "Warning", "LabTemplateNotFound", "Error")
		r.Delete(ctx, &labInstance, &client.DeleteOptions{})
		return ctrl.Result{}, err
	}
	r.EventsRecorder.Event(&labInstance, "Normal", "LabTemplateFound", "Correct")
	// prepare variables common to all resources
	name := labTemplate.Name + "-" + labInstance.Spec.StudentID
	namespace := labInstance.Namespace
	// this is added so that all resources created for this LabInstance are destroyed when the LabInstance is deleted
	labiOwnerRef := []metav1.OwnerReference{
		{
			APIVersion: labInstance.APIVersion,
			Kind:       labInstance.Kind,
			Name:       labInstance.Name,
			UID:        labInstance.UID,
		},
	}

	// 1: create secret referenced by VirtualMachineInstance (Cloudinit)
	secret := pkg.CreateSecret(name, namespace)
	secret.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		log.Info("Could not create secret " + secret.Name)
		labInstance = setLabInstanceStatus(labInstance, "SECRET ERROR", err.Error())
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else {
		log.Info("Secret " + secret.Name + " correctly created")
		labInstance = setLabInstanceStatus(labInstance, "SECRET CREATED", "")
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	}
	// 2: create pvc referenced by VirtualMachineInstance ( Persistent Data)
	// Check if exists
	// If exists, can we attach?
	// If yes, attach
	// If not, update the status with error
	pvc := pkg.CreatePerstistentVolumeClaim(name, namespace, "rook-ceph-block")
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, pvc); err != nil && err.Error() != "ALREADY EXISTS" {
		log.Info("Could not create pvc " + pvc.Name)
		labInstance = setLabInstanceStatus(labInstance, "PVC ERROR", err.Error())
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else if err != nil && err.Error() == "ALREADY EXISTS" {
		log.Info("PeristentVolumeClaim " + pvc.Name + " already exists")
		labInstance = setLabInstanceStatus(labInstance, "PVC ATTACHED", "")
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	} else {
		log.Info("PeristentVolumeClaim " + pvc.Name + " correctly created")
		labInstance = setLabInstanceStatus(labInstance, "PVC CREATED", "")
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	}

	// 3: create VirtualMachineInstance
	vmi := pkg.CreateVirtualMachineInstance(name, namespace, labTemplate, secret.Name, pvc.Name)
	vmi.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		log.Info("Could not create vm " + vmi.Name)
		labInstance = setLabInstanceStatus(labInstance, "VMI ERROR", err.Error())
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else {
		log.Info("VirtualMachineInstance " + vmi.Name + " correctly created")
		labInstance = setLabInstanceStatus(labInstance, "VMI CREATED", "")
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	}
	// 4: create Service to expose the vm
	service := pkg.CreateService(name, namespace)
	service.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		log.Info("Could not create service " + service.Name)
		labInstance = setLabInstanceStatus(labInstance, "SERVICE ERROR", err.Error())
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else {
		log.Info("Service " + service.Name + " correctly created")
		labInstance = setLabInstanceStatus(labInstance, "SERVICE CREATED", "")
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	}

	// 5: create Ingress to manage the service
	ingress := pkg.CreateIngress(name, namespace, secret.Name, service)
	ingress.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		log.Info("Could not create ingress " + ingress.Name)
		labInstance = setLabInstanceStatus(labInstance, "INGRESS ERROR", err.Error())
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else {
		log.Info("Ingress " + ingress.Name + " correctly created")
		labInstance = setLabInstanceStatus(labInstance, "INGRESS CREATED", "https://"+ingress.Spec.Rules[0].Host+"/"+name)
		if err := r.Status().Update(ctx, &labInstance); err != nil {
			log.Error(err, "unable to update LabInstance status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&instancev1.LabInstance{}).
		Complete(r)
}

func setLabInstanceStatus(labInstance instancev1.LabInstance, phase string, url string) instancev1.LabInstance {
	labInstance.Status.Phase = phase
	labInstance.Status.Url = url
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	return labInstance
}

func GetConfig(path string) (*rest.Config, error) {
	var config *rest.Config
	var err error

	if path == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else if path != "" {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			// Get the kubeconfig from the filepath.
			config, err = clientcmd.BuildConfigFromFlags("", path)
			config.GroupVersion = &virtv1.GroupVersion
			config.NegotiatedSerializer = virtv1.ne
			if err != nil {
				return nil, err
			}
		}
	}

	return config, err
}

// create a standard K8s client -> to access use client.CoreV1().<resource>(<namespace>).<method>())
func getClient(path string) (*rest.RESTClient, error) {
	config, err := GetConfig(path)
	if err != nil {
		return nil, err
	}
	return rest.RESTClientFor(config)
}

func VmWatcher(client *rest.RESTClient) error {
	resource := "virtualmachineinstances"
	namespace := corev1.NamespaceAll
	watch, err := client.Get().
		Prefix("watch").
		Namespace(namespace).
		Resource(resource).
		VersionedParams(&metav1.ListOptions{}, metav1.ParameterCodec).
		Watch()
	if err != nil {
		return err
	}

	go func() {
		for event := range watch.ResultChan() {
			vmi, ok := event.Object.(*virtv1.VirtualMachineInstance)
			if !ok {
				_ = fmt.Errorf("unexpected type")
			}
			s := vmi.Status.Phase
			print(s)
		}
	}()

	return nil
	//resp, err := http.Get("http://localhost:8001/apis/cw.com/v1/apigateways?watch=true")
	//if err != nil {
	//	panic(err)
	//}
	//defer resp.Body.Close()
	//decoder := json.NewDecoder(resp.Body)
	//for {
	//	var event v1.ApiGatewayWatchEvent
	//	if err := decoder.Decode(&event); err == io.EOF {
	//		break
	//	} else if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Printf("Received watch event: %s: %s: \n", event.Type, event.Object.Metadata.Name)
	//
	//}
	//
	//client.ExtensionsV1beta1().RESTClient().w
	//watch, err := client..CoreV1().Pods(p.config.Namespace).Watch(metav1.ListOptions{})
	//if err != nil {
	//	errors.Wrap(err, err.Error())
	//}
	//go func() {
	//	for event := range watch.ResultChan() {
	//		p2, ok := event.Object.(*v1.Pod)
	//		if !ok {
	//			_ = fmt.Errorf("unexpected type")
	//		}
	//		p.notifier(F2HTranslate(p2, p.config.RemoteNewPodCidr))
	//	}
	//}()
	//return nil
	//resp, err := http.Get(url)
	//if err != nil {
	//	panic(err)
	//}
	//defer resp.Body.Close()
	//decoder := json.NewDecoder(resp.Body)
	//go func(){
	//	for {
	//		var event metav1.WatchEvent
	//		if err := decoder.Decode(&event); err == io.EOF {
	//			break
	//		} else if err != nil {
	//
	//		}
	//
	//	}
	//}()
	//func(options metav1.ListOptions) (watch.Interface, error) {
	//	options.FieldSelector = fieldSelector.String()
	//	options.LabelSelector = labelSelector.String()
	//	return c.Get().
	//		Prefix("watch").
	//		Namespace(namespace).
	//		Resource(resource).
	//		VersionedParams(&options, metav1.ParameterCodec).
	//		Watch()
	//}

	
}
