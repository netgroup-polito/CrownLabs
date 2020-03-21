# Alert Notification
When an alert is sent to AlertManager, those are also sent to slack. 
This template distinguishes alerts based on severity and send them on the correct Slack channel.

## How to install the template
To install this template, you have to:
- add your `secret` in [alertmanager.yaml](alertmanager.yaml), encoded in `base64`;
- configure the fields `api_url` with your own Slack hook(s).

Once done, you can run the following commands:
````
$ cat <file-template> | base64 -w0
````

And then use the encoded output to insert it in the secret with this command inside the field "alertmanager.yaml"
````
$ kubectl edit secrets -n <alertmanager namespace> <alertmanager secret> -o yaml
````
