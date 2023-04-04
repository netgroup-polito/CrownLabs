## PostgreSQL-Operator
The following steps will install the postgresql-operator in the namespace called **postgres-operator**, according to [the official documentation](https://github.com/zalando/postgres-operator/blob/v1.9.0/docs/quickstart.md#helm-chart).
The Postgres Operator can be installed simply by applying `yaml` manifests, after properly changing the namespace in file `operator-service-account-rbac.yaml` for the `service account` and `cluster rolebinding`.

```sh
# add repo for postgres-operator
helm repo add postgres-operator-charts https://opensource.zalando.com/postgres-operator/charts/postgres-operator

# refresh helm repositories
helm repo update

# install the postgres-operator
helm upgrade --install postgres-operator postgres-operator-charts/postgres-operator -n postgres-operator --create-namespace -f Values.yaml
```

The currently deployed version is the version 1.9.0.

Upgrade tips:

```sh
helm show values postgres-operator-charts/postgres-operator > Values-new.yaml
```

Then, diff the Values-new.yaml and current Values.yaml. Compare and merge old customizations into the new values, then run the upgrade command above.
