# CrownLabs Image List Requestors - Guida Completa

## Riepilogo dei Cambiamenti

È stata aggiunta una nuova implementazione dell'interfaccia `Requestor` per supportare il registry **Harbor**, oltre a quello già esistente per Docker Registry V2.

### File Modificati:

1. **`operators/pkg/imageList/requestors.go`**
   - Aggiunta struct `HarborImageListRequestor` che implementa l'interfaccia `Requestor`
   - Metodi specifici per interagire con le API Harbor

2. **`operators/cmd/crownlabs-image-list/main.go`**
   - Aggiunto supporto per configurazioni multi-registry tramite file JSON
   - Mantenuta compatibilità con il modello single-registry legacy
   - Nuovi flag: `--registries-config`, `--registry-type`, `--harbor-project`

## Implementazione Harbor

### Struttura HarborImageListRequestor

```go
type HarborImageListRequestor struct {
    url         string
    username    string
    password    string
    projectName string  // Nome del progetto Harbor
    client      *http.Client
    initialized bool
    log         logr.Logger
}
```

### API Utilizzate da Harbor

- **Catalog Endpoint**: 
  ```
  GET /api/v2.0/projects/{PROJECT_NAME}/repositories?page=1&page_size=100
  ```
  Ritorna un array di oggetti repository con campi come `name`, `artifact_count`, etc.

- **Artifacts Endpoint**:
  ```
  GET /api/v2.0/projects/{PROJECT_NAME}/repositories/{REPO}/artifacts
  ```
  Ritorna un array di artifact (immagini) per un repository specifico.

## Modalità di Utilizzo

### Single Registry Mode (Legacy)

Per Docker Registry:
```bash
./crownlabs-image-list \
  --registry-type docker \
  --registry-url https://registry.example.com \
  --registry-username admin \
  --registry-password password123 \
  --advertised-registry-name registry.example.com \
  --image-list-name docker-imagelist
```

Per Harbor:
```bash
./crownlabs-image-list \
  --registry-type harbor \
  --registry-url https://harbor.example.com \
  --registry-username admin \
  --registry-password harbor123 \
  --advertised-registry-name harbor.example.com \
  --image-list-name harbor-imagelist \
  --harbor-project my-project
```

### Multi-Registry Mode (Consigliato)

```bash
./crownlabs-image-list --registries-config /path/to/config.json
```

Dove il file di configurazione contiene:

```json
{
  "registries": [
    {
      "name": "docker-internal",
      "type": "docker",
      "url": "https://registry-internal.company.com",
      "advertised": "registry-internal.company.com",
      "username": "internal_user",
      "password": "internal_password",
      "image_list_name": "docker-internal-imagelist"
    },
    {
      "name": "harbor-primary",
      "type": "harbor",
      "url": "https://harbor.company.com",
      "advertised": "harbor.company.com",
      "username": "harbor_admin",
      "password": "harbor_password",
      "image_list_name": "harbor-primary-imagelist",
      "project": "crownlabs"
    },
    {
      "name": "harbor-backup",
      "type": "harbor",
      "url": "https://harbor-backup.company.com",
      "advertised": "harbor-backup.company.com",
      "username": "harbor_admin",
      "password": "harbor_backup_password",
      "image_list_name": "harbor-backup-imagelist",
      "project": "backup-images"
    }
  ]
}
```

## Gestione della Configurazione

### SharedData per Harbor

La configurazione di Harbor utilizza `RequestersSharedData` per passare il `project_name` al `HarborImageListRequestor`:

```go
imagelist.RequestersSharedData["harbor_project_name"] = "mio-progetto"
```

Questa è un'alternativa pulita a modificare l'interfaccia `Requestor`.

## Flusso di Elaborazione

### Per Docker Registry:

1. **Fetch Catalog**: GET `/v2/_catalog` → ritorna `{"repositories": ["repo1", "repo2", ...]}`
2. **Fetch Tags**: Per ogni repo, GET `/v2/{repo}/tags/list` → ritorna `{"name": "repo", "tags": [...]}`
3. **Salva ImageList**: Salva tutti i dati come risorsa Kubernetes

### Per Harbor:

1. **Fetch Repositories**: GET `/api/v2.0/projects/{project}/repositories?page=1&page_size=100` → ritorna array di repository objects
2. **Fetch Artifacts**: Per ogni repo, GET `/api/v2.0/projects/{project}/repositories/{repo}/artifacts` → ritorna array di artifacts
3. **Salva ImageList**: Salva tutti i dati come risorsa Kubernetes

## Vantaggi della Nuova Architettura

✅ **Supporto Multi-Registry**: Aggiorna multiple sorgenti di immagini in una singola esecuzione
✅ **Configurazione Centralizzata**: Usa un file JSON per definire tutti i registry
✅ **Estensibilità**: Facile aggiungere altri tipi di registry implementando l'interfaccia `Requestor`
✅ **Error Handling Robusto**: Se un registry fallisce, gli altri continuano
✅ **ImageList Specifiche per Endpoint**: Ogni registry ha la sua propria `ImageList` Kubernetes

## Errori Comuni

### "harbor_project_name not found in RequestersSharedData"
- **Causa**: Non hai impostato `imagelist.RequestersSharedData["harbor_project_name"]`
- **Soluzione**: Usa il flag `--harbor-project` o assicurati che il file di configurazione abbia il campo `project`

### "unexpected Harbor response format"
- **Causa**: L'API Harbor ha risposto in un formato inaspettato
- **Soluzione**: Verifica che il progetto esista e che l'utente abbia le giuste permessi

### "failed to initialize K8s client"
- **Causa**: La connessione al cluster Kubernetes non è configurata
- **Soluzione**: Assicurati che kubeconfig sia correttamente impostato

## Note Importanti

1. **Pagina-Pagina nei Repository Harbor**: Il configurato endpoint usa `page=1&page_size=100`. Se hai più di 100 repository, considera di implementare un ciclo di paginazione.

2. **Autenticazione**: Harbor e Docker Registry supportano entrambi Basic Auth. Assicurati che le credenziali siano corrette.

3. **Performance**: Le richieste ai registry sono fatte in parallelo dove possibile per migliorare le prestazioni.

4. **Logging**: Usa `-v 2 -alsologtostderr` per debug logs dettagliati.

## Prossimi Passi

Se in futuro hai bisogno di aggiungere altri registry (es. Quay.io, ECR, ecc.), dovrai:

1. Creare una nuova struct che implementa `Requestor`
2. Implementare i metodi `GetImageList()` e `Initialize()`
3. Registrarla nel `init()` di `requestors.go`
4. Aggiungere il tipo nel main.go
