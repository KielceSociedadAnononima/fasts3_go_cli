# ⚡ FastS3 Go Client (`fasts3_go_cli`)

Librería cliente completa (SDK) en **Go** — **sin dependencias externas** (100% biblioteca estándar de Go) — para consumir el 100% de la API REST del servidor de almacenamiento **FastS3**.

## ✨ Características

- **Cero Dependencias Externas**: Construida exclusivamente sobre `net/http`, `context`, `encoding/json` e `io`.
- **Cobertura Total de Endpoints**: Autenticación y Bootstrap, API Keys, Usuarios, Matriz de Permisos por Bucket, Volúmenes Multi-Disco, Buckets, Archivos/Streams, Miniaturas y AutoVacuum.
- **Todos los Métodos de Autenticación**:
  - `fasts3.WithBasicAuth(user, pass)` (`Authorization: Basic ...`)
  - `fasts3.WithBearerToken(token)` (`Authorization: Bearer ...` para Bootstrap Token o API Key)
  - `fasts3.WithAPIKey(key)` (`X-API-Key: fs3_...`)
  - `fasts3.WithFastS3Token(token)` (`X-FastS3-Token: ...`)
  - `fasts3.WithCustomHeaderAuth(user, pass)` (`X-FastS3-User` + `X-FastS3-Pass`)
  - Acceso anónimo por defecto para buckets públicos
- **Streaming Real con `io.Reader` / `io.Writer`**: Subida (`UploadStream`, `UploadFile`) y descarga (`DownloadStream`, `DownloadToWriter`, `DownloadToPath`) sin cargar archivos grandes en memoria.
- **Sincronización de Carpetas / Prefijos**: `UploadDirectory`, `DownloadPrefixToDir` y `DeletePrefix`.
- **Concurrencia Segura**: `fasts3.Client` es seguro para uso concurrente entre múltiples goroutines (`sync.RWMutex`).

---

## 📦 Instalación

```bash
go get github.com/PTechSoftware/fasts3_go_cli
```

O mediante directiva `replace` local en tu `go.mod`:

```go
replace github.com/PTechSoftware/fasts3_go_cli => ../fasts3_go_cli
```

---

## 🚀 Uso Rápido

```go
package main

import (
	"context"
	"fmt"
	"log"

	fasts3 "github.com/PTechSoftware/fasts3_go_cli"
)

func main() {
	ctx := context.Background()

	// 1. Inicializar cliente con Basic Auth, Bearer Token o API Key
	cli, err := fasts3.NewClient(
		"http://localhost:8080",
		fasts3.WithBasicAuth("admin", "secret"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Verificar sesión activa
	authInfo, err := cli.VerifyAuth(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Conectado como %s (admin=%v)\n", authInfo.Username, authInfo.IsAdmin)

	// 2. Crear un bucket con cuota de 100 MB
	_, err = cli.CreateBucket(ctx, fasts3.CreateBucketRequest{
		Name:     "documentos",
		IsPublic: false,
		Quota:    100 * 1024 * 1024,
	})
	if err != nil && !fasts3.IsConflict(err) {
		log.Fatal(err)
	}

	// 3. Subir datos en memoria o archivos desde disco por streaming
	_, err = cli.UploadBytes(
		ctx,
		"documentos",
		"reportes/2026/resumen.txt",
		[]byte("Reporte anual FastS3"),
		"text/plain",
	)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Consultar detalles enriquecidos de archivos
	details, err := cli.ListFileDetails(ctx, "documentos", &fasts3.FileListFilter{
		Subpath: "reportes/2026",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range details {
		fmt.Printf("- %s (%d bytes, %s)\n", f.Key, f.Size, f.MimeType)
	}

	// 5. Descargar archivo
	raw, err := cli.DownloadBytes(ctx, "documentos", "reportes/2026/resumen.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Contenido descargado: %s\n", string(raw))

	// 6. Usuarios, Permisos y API Keys
	worker, err := cli.CreateUser(ctx, fasts3.CreateUserRequest{
		Username: "worker",
		Password: "ClaveSegura123!",
		IsAdmin:  fasts3.BoolPtr(false),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = cli.SetPermission(ctx, fasts3.SetPermissionRequest{
		UserID:     worker.ID,
		BucketName: "documentos",
		CanRead:    true,
		CanWrite:   true,
		CanDelete:  false,
	})
	if err != nil {
		log.Fatal(err)
	}

	key, err := cli.CreateAPIKey(ctx, fasts3.CreateAPIKeyRequest{
		Name:   "Worker Key",
		UserID: fasts3.Int64Ptr(worker.ID),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("API Key generada: %s\n", key.APIKey)

	// 7. Ejecutar limpieza AutoVacuum
	stats, err := cli.TriggerVacuum(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AutoVacuum completado: %d buckets reconciliados\n", stats.BucketsReconciled)
}
```

---

## 📚 Resumen de Métodos Disponibles

| Categoría | Métodos de `*fasts3.Client` |
|---|---|
| **Autenticación & Bootstrap** | `VerifyAuth`, `BootstrapStatus`, `BootstrapLogin`, `BootstrapLoginAndAuthenticate`, `BootstrapSetup`, `BootstrapSetupAndAuthenticate`, `Login`, `LoginAndAuthenticate`, `ChangePassword`, `ChangePasswordAndAuthenticate`, `GetBootstrapToken` |
| **API Keys** | `ListAPIKeys`, `CreateAPIKey`, `DeleteAPIKey` |
| **Usuarios** | `ListUsers`, `GetUserByUsername`, `CreateUser`, `UpdateUser`, `DeleteUser` |
| **Permisos RBAC** | `ListPermissions`, `SetPermission`, `DeletePermission`, `GetBucketPermissions`, `DeleteBucketPermission` |
| **Volúmenes Multi-Disco** | `ListDisks`, `AddDisk`, `DeleteDisk` |
| **Buckets** | `ListBuckets`, `GetBucketByName`, `CreateBucket`, `UpdateBucket`, `DeleteBucket`, `DeleteBucketByName` |
| **Archivos & Miniaturas** | `ListFiles`, `ListFileDetails`, `UploadBytes`, `UploadStream`, `UploadFile`, `UploadJSON`, `UploadJSONBytes`, `DownloadStream`, `DownloadWithOptions`, `DownloadFile`, `DownloadBytes`, `DownloadToWriter`, `DownloadToPath`, `GetThumbnail`, `GetThumbnailAlt`, `SaveThumbnailToPath`, `DeleteFile`, `UploadDirectory`, `DownloadPrefixToDir`, `DeletePrefix`, `FileURL`, `ThumbnailURL` |
| **AutoVacuum** | `VacuumStatus`, `TriggerVacuum` |

---

## 🧪 Ejecutar Pruebas

```bash
go vet ./...
go test -v -race ./...
```
