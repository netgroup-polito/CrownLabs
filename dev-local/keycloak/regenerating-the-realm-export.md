# Regenerating the realm export

If you manually change something in the local realm (a new client, a new scope, a new user, and so on) and you would like to freeze it for future installs, you should regenerate the export and re-wrap it into the ConfigMap manifest:

```bash
export KUBECONFIG=~/.kube/config

# This runs the export while the pod is already running. kc.sh runs as a process
# separate from the main server, so you do not need to stop Keycloak for a simple
# local snapshot.
kubectl exec keycloak-0 -- rm -rf /tmp/export
kubectl exec keycloak-0 -- /opt/keycloak/bin/kc.sh export --dir /tmp/export --realm crownlabs --users realm_file

# Copy the file out of the pod. The image has no "tar", so "kubectl cp" will not work:
# use "cat" instead. Then wrap the file into the ConfigMap manifest committed in this folder.
kubectl exec keycloak-0 -- cat /tmp/export/crownlabs-realm.json > /tmp/crownlabs-realm.json
{
  echo "# Generated from a \`kc.sh export\` of the crownlabs realm — do not hand-edit."
  echo "# See regenerating-the-realm-export.md to refresh this file."
  kubectl create configmap keycloak-realm-import \
    --from-file=crownlabs-realm.json=/tmp/crownlabs-realm.json --dry-run=client -o yaml
} > dev-local/keycloak/manifests/crownlabs-realm-configmap.yaml
rm /tmp/crownlabs-realm.json
```

> **Note**: `kc.sh export` exits with an error (`Address already in use` on port 9000).
> This happens because the export process tries to start the management interface on the same port the running main server already uses.
> **The error is harmless, and happens after the data export has already completed.**
> Check the logs for `KC-SERVICES0035: Export finished successfully` before the error, to confirm this.
> For an export without this warning, stop Keycloak first (`kubectl scale statefulset keycloak --replicas=0`), as recommended by the [official documentation](https://www.keycloak.org/server/importExport).

**Before you commit a new export**, always check that any newly auto-generated Keycloak client secret is pinned to a known value, the same way `operator-local-dev-secret` is, instead of ending up in the repository as a random string:

```bash
TOKEN=$(curl -sk -X POST https://keycloak.crownlabs.local/realms/master/protocol/openid-connect/token \
  -d "client_id=admin-cli" -d "username=admin" -d "password=admin" -d "grant_type=password" \
  | python3 -c "import json,sys; print(json.load(sys.stdin)['access_token'])")
CLIENT_UUID=$(curl -sk "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients?clientId=<client-id>" \
  -H "Authorization: Bearer $TOKEN" | python3 -c "import json,sys; print(json.load(sys.stdin)[0]['id'])")
curl -sk -X PUT "https://keycloak.crownlabs.local/admin/realms/crownlabs/clients/$CLIENT_UUID" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"secret": "<your-chosen-known-value>"}'
```

If the cluster is already running, and you just want to push a freshly regenerated export without reinstalling everything:

```bash
kubectl apply -f dev-local/keycloak/manifests/crownlabs-realm-configmap.yaml
```

`--import-realm` **does not overwrite** an already-existing realm.
This command only updates the ConfigMap for future installs, or for after you delete the existing realm.
