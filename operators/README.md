# CrownLabs operators

## LabInstance
The commands below are written assuming your working directory is `labInstance-operator`.

### LabTemplate
For a LabInstance to exist the corresponding LabTemplate must be already present in the cluster.
The steps necessary to modify and install a LabTemplate resource are the same of LabInstance.

To modify the LabTemplate CRD you need to
1. open the file _labTemplate/api/v1/labtemplate_types.go_. Here you have all the `struct` types related to the LabTemplate CRD
2. add/modify/delete the fields of `LabTemplateSpec` and/or `LabTemplateStatus`
3. run `make lab-template`; this will regenerate the code for the new version of LabTemplate
4. you can find the CRD generated in _labTemplate_crd_bases_

To install the LabTemplate CRD on your cluster
1. run `make install-lab-template`. This will install the CRD LabTemplate on your cluster.
2. in _labTemplate/samples_ you have an example of LabTemplate object. If you want to create it run `kubectl apply -f labTemplate/samples/template_v1_labtemplate.yaml`

If you want to delete the CRD run `make uninstall-lab-template`.

### CRD generation
To modify the LabInstance CRD you need to
1. open the file _labInstance-operator/api/v1/labinstance_types.go_. Here you have all the `struct` types related to the LabInstance CRD
2. add/modify/delete the fields of `LabInstanceSpec` and/or `LabInstanceStatus`
3. run `make`; this will regenerate the code for the new version of LabInstance
4. you can find the CRD generate in _config_crd_bases_

### Installation
1. run `make install`. This will install the CRD LabInstance on your cluster.
2. in _config/samples_ you have an example of LabInstance object. If you want to create it run `kubectl apply -f config/samples/instance_v1_labinstance.yaml`
3. you can get the list of CRD installed on your cluster by running `kubectl get crd`.
4. To get the list of LabInstance resources run `kubectl get labi`.

If you want to delete the CRD run `make uninstall`.

### Controller logic
The logic of the controller should be put under _controllers/labinstance_controller.go_, in the `Reconcile` method.
When a LabInstance resource is created, the `Reconcile` method is triggered. 
It is checked if a LabTemplate resource with the associated name exists, and in this case the `Status` of the LabInstance is set to "DEPLOYED".

### Run instructions
To run the application

- **outside** a cluster: run 'make run'