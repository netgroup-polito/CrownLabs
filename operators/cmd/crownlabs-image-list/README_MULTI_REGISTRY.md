# CrownLabs Image List Updater

## Utilizzo

Il tool `crownlabs-image-list` è responsabile di aggiornare le risorse Kubernetes `ImageList` con l'elenco delle immagini disponibili nei registry configurati.

### Modalità 1: Single Registry (Legacy)

Per aggiornare un singolo registry, usare i seguenti flag:

```bash
./crownlabs-image-list \
  --registry-type docker \
  --registry-url https://registry.example.com \
  --registry-username admin \
  --registry-password password \
  --advertised-registry-name registry.example.com \
  --image-list-name docker-imagelist
```

Per Harbor, aggiungere il flag `--harbor-project`:

```bash
./crownlabs-image-list \
  --registry-type harbor \
  --registry-url https://harbor.example.com \
  --registry-username admin \
  --registry-password password \
  --advertised-registry-name harbor.example.com \
  --image-list-name harbor-imagelist \
  --harbor-project crownlabs-project
```

### Modalità 2: Multi-Registry (Consigliata)

Per aggiornare più registry contemporaneamente, usare un file di configurazione JSON:

```bash
./crownlabs-image-list \
  --registries-config /path/to/registries-config.json
```

#### Formato del file di configurazione

```json
{
  "registries": [
    {
      "name": "docker-registry",
      "type": "docker",
      "url": "https://registry.example.com",
      "advertised": "registry.example.com",
      "username": "admin",
      "password": "password",
      "image_list_name": "docker-imagelist"
    },
    {
      "name": "harbor-registry",
      "type": "harbor",
      "url": "https://harbor.example.com",
      "advertised": "harbor.example.com",
      "username": "admin",
      "password": "password",
      "image_list_name": "harbor-imagelist",
      "project": "project-name"
    }
  ]
}
```

#### Campi della configurazione

- **name**: Nome del registry (usato per il logging)
- **type**: Tipo di registry ("docker" o "harbor")
- **url**: URL base del registry (es. https://registry.example.com)
- **advertised**: Hostname del registry come propagato ai consumatori
- **username**: Username per l'autenticazione
- **password**: Password per l'autenticazione
- **image_list_name**: Nome della risorsa Kubernetes ImageList dove salvare i dati
- **project** (solo Harbor): Nome del progetto Harbor da cui recuperare i repository

## Implementazioni supportate

### Docker Registry V2

Il requestor `DefaultImageListRequestor` supporta i registry Docker compatibili con l'API V2 standard.

**API utilizzate:**
- `GET /v2/_catalog` - Per ottenere l'elenco dei repository
- `GET /v2/{repository}/tags/list` - Per ottenere i tag di ogni repository

### Harbor

Il requestor `HarborImageListRequestor` supporta il registry Harbor.

**API utilizzate:**
- `GET /api/v2.0/projects/{project}/repositories?page=1&page_size=100` - Per ottenere l'elenco dei repository
- `GET /api/v2.0/projects/{project}/repositories/{repository}/artifacts` - Per ottenere gli artifact di ogni repository

## Note importanti

1. **Validazione della configurazione**: Se si usa il file di configurazione multi-registry, l'applicazione continuerà a processare gli altri registry anche se uno fallisce.

2. **Credenziali sensibili**: Le credenziali possono essere fornite tramite variabili di ambiente anche se passate nel file di configurazione (è consigliato non hardcodare le password nei file di configurazione).

3. **Logging**: Usare i flag `-v` or `-alsologtostderr` per view il logging dettagliato.

```bash
./crownlabs-image-list \
  --registries-config registries-config.json \
  -v 2 -alsologtostderr
```

## Esempi di utilizzo

### Esempio 1: Aggiornare un singolo registry Docker

```bash
./crownlabs-image-list \
  --registry-type docker \
  --registry-url https://docker.example.com \
  --registry-username myuser \
  --registry-password mypassword \
  --advertised-registry-name docker.example.com \
  --image-list-name primary-docker-registry
```

### Esempio 2: Aggiornare un singolo registry Harbor

```bash
./crownlabs-image-list \
  --registry-type harbor \
  --registry-url https://harbor.example.com \
  --registry-username admin \
  --registry-password harborcredentials \
  --advertised-registry-name harbor.example.com \
  --image-list-name harbor-crownlabs \
  --harbor-project crownlabs
```

### Esempio 3: Aggiornare più registry usando configurazione JSON

```bash
./crownlabs-image-list \
  --registries-config /etc/crownlabs/registries-config.json
```

Dove `/etc/crownlabs/registries-config.json` contiene:

```json
{
  "registries": [
    {
      "name": "internal-docker",
      "type": "docker",
      "url": "https://registry-internal.example.com",
      "advertised": "registry-internal.example.com",
      "username": "internal-user",
      "password": "internal-pass",
      "image_list_name": "internal-docker-images"
    },
    {
      "name": "public-harbor",
      "type": "harbor",
      "url": "https://harbor-public.example.com",
      "advertised": "harbor-public.example.com",
      "username": "public-user",
      "password": "public-pass",
      "image_list_name": "public-harbor-images",
      "project": "public-project"
    },
    {
      "name": "backup-harbor",
      "type": "harbor",
      "url": "https://harbor-backup.example.com",
      "advertised": "harbor-backup.example.com",
      "username": "backup-user",
      "password": "backup-pass",
      "image_list_name": "backup-harbor-images",
      "project": "backup-project"
    }
  ]
}
```
