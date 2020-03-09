# CrownLabs operators

##LabTemplate
The commands below are written assuming your working directory is `labTemplate-operator`.

### CRD generation
To modify the LabTemplate CRD you need to
1. open the file _labTemplate-operator/api/v1/labtemplate_types.go_. Here you have all the `struct` types related to the LabTemplate CRD
2. add/modify/delete the fields of `LabTemplateSpec` and/or `LabTemplateStatus`
3. run `make`; this will regenerate the code for the new version of LabTemplate

### Installation
1. run `make install`. This will install the CRD LabTemplate on your cluster.
2. in _config/samples_ you have an example of LabTemplate object. If you want to create it run `kubectl apply -f config/samples/template_v1_labtemplate.yaml`
3. you can get the list of CRD installed on your cluster by running `kubectl get crd`.
To get the list of LabTemplate resources run `kubectl get labt` (`labi` for LabInstance).

### Controller logic
The logic of the controller should be put under _controllers/labtemplate_controller.go_, in the `Reconcile` method.

##LabInstance

The instructions for LabInstance generation and installation are exactly the same of LabTemplate, you just need to replace _labTemplate_ with _labInstance_. 

### Operation
When a LabInstance resource is created, the `Reconcile` method is triggered. It is checked if a LabTemplate resource with the associated name exists, and in this case the `Status` of the LabInstance is set to "DEPLOYED".