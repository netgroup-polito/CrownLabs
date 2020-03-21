# Alert Notification
When an alert is sent to AlertManager, those are also sent to slack. 
This template distinguishes alerts based on severity and send it to the correct channel.
## How to install the tamplate

To be installed this template, it must be entered in the alertmanager.yaml field inside the secret of the alertmanager, encoded in base64.
Before to do this the field "api_url" in the tamplate must be updated with your own slack hook.
After that, you can run the following commands:

````
$ cat <file-template> | base64 -w0
````
And then use encoded output to insert it in the secret with this command inside the field "alertmanager.yaml"
````
$ kubectl edit secrets -n <alertmanager namespace> <alertmanager secret> -o yaml
````
