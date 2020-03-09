# LabTemplate Operator

K8s operator that allows to create and upload a "template" lab, e.g., a VM installed with the proper software, which will be instantiated multiple times and connected to its associated user.

Instantiated templates will be handled by the [LabInstance](../labInstance-operator) operator.
