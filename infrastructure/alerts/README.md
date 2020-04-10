# Alert Notification
When an alert is sent to AlertManager, those are also sent to Slack. 
This template differentiates alerts based on severity and sends them on the correct Slack channel.

## How to install the template
To install this template, you have to follow the steps below:

1) Configure the fields `api_url` in [alertmanager.yaml](alertmanager.yaml) with your own Slack hook(s).

2) Then encode the above template in base64 (in our case, `<file-template>` is `alertmanager.yaml`):
````
$ cat <file-template> | base64 -w0
````

3) Now, you have to edit the secrets of your alertmanager deployment and add the above output (i.e., the entire content of `alertmanager.yaml`, encoded in based 64) as a *secret* in correspondence of field `alertmanager.yaml`. The above command will open an editor that will allow to complete this action:
````
$ kubectl edit secrets -n <alertmanager namespace> <alertmanager secret name> -o yaml
````
