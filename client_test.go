package fasts3_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	fasts3 "github.com/PTechSoftware/fasts3_go_cli"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestAuthUsersPermissionsDisksAPIKeysAndVacuum(t *testing.T) {
	var mu sync.Mutex
	var seenPaths []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seenPaths = append(seenPaths, r.Method+" "+r.URL.Path)
		mu.Unlock()

		switch r.Method + " " + r.URL.Path {
		case "GET /api/auth/bootstrap-status":
			writeJSON(w, http.StatusOK, map[string]any{
				"is_configured": false,
				"message":       "Pending initial setup",
			})
		case "POST /api/auth/bootstrap-login":
			writeJSON(w, http.StatusOK, map[string]any{
				"status":   "authenticated",
				"is_admin": true,
				"username": "bootstrap_admin",
				"token":    "fasts3_sec_boot_999",
			})
		case "POST /api/auth/bootstrap-setup":
			writeJSON(w, http.StatusOK, map[string]any{
				"status":     "configured",
				"message":    "Admin configured",
				"username":   "admin",
				"user_id":    1,
				"is_admin":   true,
				"basic_auth": "Basic YWRtaW46U3VwZXJQYXNzIQ==",
			})
		case "POST /api/auth/login":
			writeJSON(w, http.StatusOK, map[string]any{
				"status":     "authenticated",
				"is_admin":   true,
				"username":   "admin",
				"basic_auth": "Basic YWRtaW46U3VwZXJQYXNzIQ==",
			})
		case "POST /api/auth/change-password":
			writeJSON(w, http.StatusOK, map[string]any{
				"status":     "success",
				"message":    "Password updated",
				"username":   "admin",
				"basic_auth": "Basic YWRtaW46TmV3UGFzcyEyMw==",
			})
		case "GET /api/auth/verify":
			writeJSON(w, http.StatusOK, map[string]any{
				"status":       "authenticated",
				"is_admin":     true,
				"username":     "admin",
				"user_id":      1,
				"is_bootstrap": false,
			})
		case "GET /api/auth/bootstrap-token":
			writeJSON(w, http.StatusOK, map[string]any{
				"bootstrap_token": "fasts3_sec_boot_999",
			})
		case "POST /api/users":
			writeJSON(w, http.StatusCreated, map[string]any{
				"id":         2,
				"username":   "bob",
				"is_admin":   false,
				"created_at": "2026-11-11T12:00:00Z",
			})
		case "GET /api/users":
			writeJSON(w, http.StatusOK, []map[string]any{
				{"id": 1, "username": "admin", "is_admin": true, "created_at": "2026-11-11T11:00:00Z"},
				{"id": 2, "username": "bob", "is_admin": false, "created_at": "2026-11-11T12:00:00Z"},
			})
		case "PUT /api/users/2":
			writeJSON(w, http.StatusOK, map[string]any{
				"id":         2,
				"username":   "bob",
				"is_admin":   true,
				"created_at": "2026-11-11T12:00:00Z",
			})
		case "DELETE /api/users/2":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "User deleted successfully"})
		case "POST /api/api-keys":
			writeJSON(w, http.StatusCreated, map[string]any{
				"id":         15,
				"user_id":    2,
				"username":   "bob",
				"is_admin":   false,
				"name":       "Prod Key",
				"key_prefix": "fs3_xyz",
				"api_key":    "fs3_xyz987654321",
				"created_at": "2026-11-11T12:05:00Z",
			})
		case "GET /api/api-keys":
			writeJSON(w, http.StatusOK, []map[string]any{
				{
					"id":         15,
					"user_id":    2,
					"username":   "bob",
					"is_admin":   false,
					"name":       "Prod Key",
					"key_prefix": "fs3_xyz",
					"created_at": "2026-11-11T12:05:00Z",
				},
			})
		case "DELETE /api/api-keys/15":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "API Key revocada exitosamente"})
		case "POST /api/permissions":
			writeJSON(w, http.StatusOK, map[string]any{
				"id":          7,
				"user_id":     2,
				"username":    "bob",
				"bucket_name": "backups",
				"can_read":    true,
				"can_write":   true,
				"can_delete":  false,
			})
		case "GET /api/permissions":
			writeJSON(w, http.StatusOK, []map[string]any{{
				"id": 7, "user_id": 2, "username": "bob", "bucket_name": "backups",
				"can_read": true, "can_write": true, "can_delete": false,
			}})
		case "GET /api/buckets/backups/permissions":
			writeJSON(w, http.StatusOK, []map[string]any{{
				"id": 7, "user_id": 2, "username": "bob", "bucket_name": "backups",
				"can_read": true, "can_write": true, "can_delete": false,
			}})
		case "DELETE /api/buckets/backups/permissions/2":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "Bucket permission removed successfully"})
		case "DELETE /api/permissions/7":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "Permission removed successfully"})
		case "POST /api/disks":
			writeJSON(w, http.StatusCreated, map[string]any{
				"id":                "ssd_2",
				"name":              "Fast SSD",
				"path":              "/mnt/ssd2",
				"is_default":        false,
				"created_at":        "2026-11-11T12:10:00Z",
				"total_space":       500000000,
				"available_space":   400000000,
				"used_space":        100000000,
				"fasts3_used_space": 8192,
			})
		case "GET /api/disks":
			writeJSON(w, http.StatusOK, []map[string]any{{
				"id":                "ssd_2",
				"name":              "Fast SSD",
				"path":              "/mnt/ssd2",
				"is_default":        false,
				"created_at":        "2026-11-11T12:10:00Z",
				"total_space":       500000000,
				"available_space":   400000000,
				"used_space":        100000000,
				"fasts3_used_space": 8192,
			}})
		case "DELETE /api/disks/ssd_2":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "Disk deleted successfully"})
		case "GET /api/vacuum/status", "POST /api/vacuum":
			writeJSON(w, http.StatusOK, map[string]any{
				"files_deleted":      4,
				"bytes_reclaimed":    32768,
				"buckets_reconciled": 3,
				"last_run_timestamp": 1794400500,
				"duration_ms":        9,
			})
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "Not found", "status": 404})
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	cli, err := fasts3.NewClient(srv.URL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	st, err := cli.BootstrapStatus(ctx)
	if err != nil || st.IsConfigured {
		t.Fatalf("BootstrapStatus: err=%v configured=%v", err, st)
	}

	bootResp, err := cli.BootstrapLoginAndAuthenticate(ctx, "fasts3_sec_boot_999")
	if err != nil || !bootResp.IsAdmin {
		t.Fatalf("BootstrapLoginAndAuthenticate: %v", err)
	}
	if cli.Credentials().Kind != fasts3.AuthBearer {
		t.Fatalf("expected AuthBearer after bootstrap login")
	}

	setupResp, err := cli.BootstrapSetupAndAuthenticate(ctx, fasts3.BootstrapSetupRequest{
		Token:    "fasts3_sec_boot_999",
		Username: "admin",
		Password: "SuperPass!",
	})
	if err != nil || setupResp.UserID != 1 {
		t.Fatalf("BootstrapSetupAndAuthenticate: %v", err)
	}
	if cli.Credentials().Password != "SuperPass!" {
		t.Fatalf("expected Basic credentials updated")
	}

	if _, err := cli.LoginAndAuthenticate(ctx, "admin", "SuperPass!"); err != nil {
		t.Fatalf("LoginAndAuthenticate: %v", err)
	}

	if _, err := cli.ChangePasswordAndAuthenticate(ctx, fasts3.ChangePasswordRequest{
		CurrentPassword: fasts3.StringPtr("SuperPass!"),
		NewPassword:     "NewPass!23",
	}); err != nil {
		t.Fatalf("ChangePasswordAndAuthenticate: %v", err)
	}
	if cli.Credentials().Password != "NewPass!23" {
		t.Fatalf("expected rotated password NewPass!23, got %q", cli.Credentials().Password)
	}

	ver, err := cli.VerifyAuth(ctx)
	if err != nil || !ver.IsAdmin {
		t.Fatalf("VerifyAuth: %v", err)
	}

	tok, err := cli.GetBootstrapToken(ctx)
	if err != nil || tok == nil || *tok != "fasts3_sec_boot_999" {
		t.Fatalf("GetBootstrapToken: %v", err)
	}

	bob, err := cli.CreateUser(ctx, fasts3.CreateUserRequest{
		Username: "bob",
		Password: "bobpass",
		IsAdmin:  fasts3.BoolPtr(false),
	})
	if err != nil || bob.ID != 2 {
		t.Fatalf("CreateUser: %v", err)
	}

	foundBob, err := cli.GetUserByUsername(ctx, "bob")
	if err != nil || foundBob == nil || foundBob.ID != 2 {
		t.Fatalf("GetUserByUsername: %v", err)
	}

	updatedBob, err := cli.UpdateUser(ctx, 2, fasts3.UpdateUserRequest{IsAdmin: fasts3.BoolPtr(true)})
	if err != nil || !updatedBob.IsAdmin {
		t.Fatalf("UpdateUser: %v", err)
	}

	createdKey, err := cli.CreateAPIKey(ctx, fasts3.CreateAPIKeyRequest{
		Name:   "Prod Key",
		UserID: fasts3.Int64Ptr(2),
	})
	if err != nil || !strings.HasPrefix(createdKey.APIKey, "fs3_") {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	keys, err := cli.ListAPIKeys(ctx)
	if err != nil || len(keys) != 1 {
		t.Fatalf("ListAPIKeys: %v", err)
	}

	if _, err := cli.DeleteAPIKey(ctx, 15); err != nil {
		t.Fatalf("DeleteAPIKey: %v", err)
	}

	perm, err := cli.SetPermission(ctx, fasts3.SetPermissionRequest{
		UserID:     2,
		BucketName: "backups",
		CanRead:    true,
		CanWrite:   true,
		CanDelete:  false,
	})
	if err != nil || perm.ID != 7 {
		t.Fatalf("SetPermission: %v", err)
	}

	if perms, err := cli.ListPermissions(ctx); err != nil || len(perms) != 1 {
		t.Fatalf("ListPermissions: %v", err)
	}
	if bPerms, err := cli.GetBucketPermissions(ctx, "backups"); err != nil || len(bPerms) != 1 {
		t.Fatalf("GetBucketPermissions: %v", err)
	}
	if _, err := cli.DeleteBucketPermission(ctx, "backups", 2); err != nil {
		t.Fatalf("DeleteBucketPermission: %v", err)
	}
	if _, err := cli.DeletePermission(ctx, 7); err != nil {
		t.Fatalf("DeletePermission: %v", err)
	}
	if _, err := cli.DeleteUser(ctx, 2); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	disk, err := cli.AddDisk(ctx, fasts3.CreateDiskRequest{
		ID:   fasts3.StringPtr("ssd_2"),
		Name: "Fast SSD",
		Path: "/mnt/ssd2",
	})
	if err != nil || disk.ID != "ssd_2" {
		t.Fatalf("AddDisk: %v", err)
	}
	if disks, err := cli.ListDisks(ctx); err != nil || len(disks) != 1 {
		t.Fatalf("ListDisks: %v", err)
	}
	if _, err := cli.DeleteDisk(ctx, "ssd_2"); err != nil {
		t.Fatalf("DeleteDisk: %v", err)
	}

	if vs, err := cli.VacuumStatus(ctx); err != nil || vs.FilesDeleted != 4 {
		t.Fatalf("VacuumStatus: %v", err)
	}
	if vr, err := cli.TriggerVacuum(ctx); err != nil || vr.BytesReclaimed != 32768 {
		t.Fatalf("TriggerVacuum: %v", err)
	}

	mu.Lock()
	count := len(seenPaths)
	mu.Unlock()
	if count < 20 {
		t.Fatalf("expected >= 20 recorded requests, got %d", count)
	}
}

func TestBucketsFilesStreamingThumbnailsDirectoriesAndErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "fs3_live_key_777" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "Unauthorized", "status": 401})
			return
		}

		switch r.Method + " " + r.URL.EscapedPath() {
		case "POST /api/buckets":
			writeJSON(w, http.StatusCreated, map[string]any{"status": "success", "message": "Bucket created successfully"})
		case "GET /api/buckets":
			writeJSON(w, http.StatusOK, []map[string]any{{
				"id":        99,
				"name":      "assets",
				"is_public": true,
				"quota":     10485760,
				"used":      256,
			}})
		case "PUT /api/buckets/99":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "Bucket updated successfully"})
		case "DELETE /api/buckets/99/assets":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "Bucket deleted successfully"})
		case "PUT /api/buckets/assets/files/docs/guia%202026.txt",
			"PUT /api/buckets/assets/files/sync/a.txt",
			"PUT /api/buckets/assets/files/sync/sub/b.txt":
			writeJSON(w, http.StatusCreated, map[string]any{"status": "success", "message": "File uploaded successfully"})
		case "POST /api/buckets/assets/files/upload":
			writeJSON(w, http.StatusCreated, map[string]any{"status": "success", "message": "File uploaded successfully"})
		case "GET /api/buckets/assets/files":
			if r.URL.Query().Get("subpath") == "sync" {
				writeJSON(w, http.StatusOK, []string{"sync/a.txt", "sync/sub/b.txt"})
				return
			}
			writeJSON(w, http.StatusOK, []string{"docs/guia 2026.txt"})
		case "GET /api/buckets/assets/files/details":
			writeJSON(w, http.StatusOK, []map[string]any{{
				"key":       "docs/guia 2026.txt",
				"name":      "guia 2026.txt",
				"size":      12,
				"mime_type": "text/plain",
				"extension": "txt",
				"is_image":  false,
			}})
		case "GET /api/buckets/assets/files/docs/guia%202026.txt",
			"GET /api/buckets/assets/files/sync/a.txt",
			"GET /api/buckets/assets/files/sync/sub/b.txt":
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Disposition", `inline; filename="guia 2026.txt"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Hola FastS3!"))
		case "GET /api/buckets/assets/files/img/banner.png/thumbnail",
			"GET /api/buckets/assets/thumbnail/img/banner.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte{0x89, 'P', 'N', 'G'})
		case "DELETE /api/buckets/assets/files/docs/guia%202026.txt",
			"DELETE /api/buckets/assets/files/sync/a.txt",
			"DELETE /api/buckets/assets/files/sync/sub/b.txt":
			writeJSON(w, http.StatusOK, map[string]any{"status": "success", "message": "File deleted successfully"})
		case "PUT /api/buckets/assets/files/quota_fail.bin":
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "Storage quota exceeded", "status": 403})
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "Not found", "status": 404})
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	cli, err := fasts3.NewClient(srv.URL, fasts3.WithAPIKey("fs3_live_key_777"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := cli.CreateBucket(ctx, fasts3.CreateBucketRequest{
		Name:     "assets",
		IsPublic: true,
		Quota:    10485760,
	}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	bkt, err := cli.GetBucketByName(ctx, "assets")
	if err != nil || bkt == nil || bkt.ID != 99 {
		t.Fatalf("GetBucketByName: %v", err)
	}

	if _, err := cli.UpdateBucket(ctx, bkt.ID, fasts3.UpdateBucketRequest{
		Name:     bkt.Name,
		IsPublic: true,
		Quota:    20971520,
	}); err != nil {
		t.Fatalf("UpdateBucket: %v", err)
	}

	if _, err := cli.UploadBytes(ctx, "assets", "docs/guia 2026.txt", []byte("Hola FastS3!"), "text/plain"); err != nil {
		t.Fatalf("UploadBytes: %v", err)
	}

	if _, err := cli.UploadJSONBytes(ctx, "assets", "docs/guia 2026.txt", fasts3.StringPtr("guia 2026.txt"), []byte("Hola FastS3!")); err != nil {
		t.Fatalf("UploadJSONBytes: %v", err)
	}

	details, err := cli.ListFileDetails(ctx, "assets", &fasts3.FileListFilter{Subpath: "docs", Name: "guia"})
	if err != nil || len(details) != 1 || details[0].Size != 12 {
		t.Fatalf("ListFileDetails: %v", err)
	}

	dl, err := cli.DownloadFile(ctx, "assets", "docs/guia 2026.txt")
	if err != nil || string(dl.Data) != "Hola FastS3!" || dl.Filename != "guia 2026.txt" {
		t.Fatalf("DownloadFile: err=%v dl=%+v", err, dl)
	}

	tmpDir := t.TempDir()
	localOut := filepath.Join(tmpDir, "downloads", "guia.txt")
	n, err := cli.DownloadToPath(ctx, "assets", "docs/guia 2026.txt", localOut)
	if err != nil || n != 12 {
		t.Fatalf("DownloadToPath: err=%v n=%d", err, n)
	}

	// Directory upload / prefix download / prefix delete
	srcDir := filepath.Join(tmpDir, "src_sync")
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "b.txt"), []byte("B"), 0o644); err != nil {
		t.Fatal(err)
	}

	uploadedKeys, err := cli.UploadDirectory(ctx, "assets", "sync", srcDir)
	if err != nil || len(uploadedKeys) != 2 {
		t.Fatalf("UploadDirectory: err=%v keys=%v", err, uploadedKeys)
	}

	pulledDir := filepath.Join(tmpDir, "pulled")
	pulledPaths, err := cli.DownloadPrefixToDir(ctx, "assets", "sync", pulledDir)
	if err != nil || len(pulledPaths) != 2 {
		t.Fatalf("DownloadPrefixToDir: err=%v paths=%v", err, pulledPaths)
	}

	deletedCount, err := cli.DeletePrefix(ctx, "assets", "sync")
	if err != nil || deletedCount != 2 {
		t.Fatalf("DeletePrefix: err=%v count=%d", err, deletedCount)
	}

	// Thumbnails
	thumbOpts := &fasts3.ThumbnailOptions{Width: 320, Height: 240}
	thumb, err := cli.GetThumbnail(ctx, "assets", "img/banner.png", thumbOpts)
	if err != nil || thumb.ContentType != "image/png" || len(thumb.Data) != 4 {
		t.Fatalf("GetThumbnail: %v", err)
	}
	thumbAlt, err := cli.GetThumbnailAlt(ctx, "assets", "img/banner.png", thumbOpts)
	if err != nil || len(thumbAlt.Data) != 4 {
		t.Fatalf("GetThumbnailAlt: %v", err)
	}

	// Direct URLs
	fURL, err := cli.FileURL("assets", "docs/guia 2026.txt", true)
	if err != nil || !strings.HasSuffix(fURL, "/api/buckets/assets/files/docs/guia%202026.txt?api_key=fs3_live_key_777") {
		t.Fatalf("FileURL: err=%v url=%s", err, fURL)
	}
	tURL, err := cli.ThumbnailURL("assets", "img/banner.png", thumbOpts, true)
	if err != nil || !strings.Contains(tURL, "width=320&height=240&api_key=fs3_live_key_777") {
		t.Fatalf("ThumbnailURL: err=%v url=%s", err, tURL)
	}

	// Quota exceeded error check
	_, qErr := cli.UploadBytes(ctx, "assets", "quota_fail.bin", []byte("123"), "")
	if !fasts3.IsForbidden(qErr) || !fasts3.IsQuotaExceeded(qErr) {
		t.Fatalf("expected quota exceeded error, got %v", qErr)
	}

	if _, err := cli.DeleteFile(ctx, "assets", "docs/guia 2026.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if _, err := cli.DeleteBucketByName(ctx, "assets"); err != nil {
		t.Fatalf("DeleteBucketByName: %v", err)
	}
}
