# Alert Notification
When an alert is sent to AlertManager, those are also sent to slack. 
This template distinguishes alerts based on severity and send them on the correct Slack channel.

## How to install the template
To install this template, first of all you have to configure the fields `api_url` with your own Slack hook(s). Then encode the template in base64 and insert it in alertmanager secret in the field alermanager.yaml.


Once done, you can run the following commands:
````
$ cat <file-template> | base64 -w0
````

And then use the encoded output to insert it in the secret with this command inside the field "alertmanager.yaml"
````
$ kubectl edit secrets -n <alertmanager namespace> <alertmanager secret name> -o yaml
````
