 
## SETUP INFORMATION
In order to create the ingress controller first of all there is a mandatory file to be applied with the deployment of the ingress controller:

```sh
kubectl apply -f mandatory.yaml
```

After that the LoadBalancer service for the ingress controller can be created

```sh
kubectl apply -f ingress_nginx.yaml
```

Once the service is created we must check what IP have been assigned to it using the command 

```sh
kubectl get svc -n ingress-nginx -o wide
```

The output will be similar to the following one 

```sh
NAME            TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)                                     AGE
ingress-nginx   LoadBalancer   10.104.98.160   192.168.31.136   80:31718/TCP,443:30654/TCP,4443:30423/TCP   60m
```

Before applying the ingress rules check that address in the ingress.yaml file is the one shown by the previous command (192.168.31.136). Otherwise all the occurrencies in the ingress.yaml need to be replaced (never forget to update the DNS records)

```sh
  ...
  status:
    loadBalancer:
      ingress:
      - ip: 192.168.31.136
   ...    
```

Once everithing is checked we can apply the ingress rules

```sh
kubectl apply -f ingress.yaml
```
