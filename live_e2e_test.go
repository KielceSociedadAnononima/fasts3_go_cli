package fasts3_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	fasts3 "github.com/PTechSoftware/fasts3_go_cli"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0x18, 0xDD, 0x8D, 0xB0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
	0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestLiveFastS3ServerEndToEnd(t *testing.T) {
	liveURL := os.Getenv("FASTS3_LIVE_URL")
	if liveURL == "" {
		t.Skip("FASTS3_LIVE_URL not set; skipping live server E2E test")
	}
	bootToken := os.Getenv("FASTS3_LIVE_BOOTSTRAP_TOKEN")
	if bootToken == "" {
		t.Fatal("FASTS3_LIVE_BOOTSTRAP_TOKEN must be set when FASTS3_LIVE_URL is set")
	}

	ctx := context.Background()
	tempRoot := t.TempDir()
	extraDiskPath := filepath.Join(tempRoot, "disk_go")
	if err := os.MkdirAll(extraDiskPath, 0o755); err != nil {
		t.Fatal(err)
	}

	admin, err := fasts3.NewClient(liveURL)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	// 1. Bootstrap status, login, setup, login, change password, verify
	if _, err := admin.BootstrapStatus(ctx); err != nil {
		t.Fatalf("BootstrapStatus: %v", err)
	}
	if _, err := admin.BootstrapLoginAndAuthenticate(ctx, bootToken); err != nil {
		t.Fatalf("BootstrapLoginAndAuthenticate: %v", err)
	}
	if _, err := admin.BootstrapSetupAndAuthenticate(ctx, fasts3.BootstrapSetupRequest{
		Token:    bootToken,
		Username: "go_admin",
		Password: "GoPass123!",
	}); err != nil {
		t.Fatalf("BootstrapSetupAndAuthenticate: %v", err)
	}
	if _, err := admin.LoginAndAuthenticate(ctx, "go_admin", "GoPass123!"); err != nil {
		t.Fatalf("LoginAndAuthenticate: %v", err)
	}
	if _, err := admin.ChangePasswordAndAuthenticate(ctx, fasts3.ChangePasswordRequest{
		CurrentPassword: fasts3.StringPtr("GoPass123!"),
		NewPassword:     "GoPass456!",
	}); err != nil {
		t.Fatalf("ChangePasswordAndAuthenticate: %v", err)
	}

	ver, err := admin.VerifyAuth(ctx)
	if err != nil || !ver.IsAdmin || ver.Username != "go_admin" {
		t.Fatalf("VerifyAuth: err=%v ver=%+v", err, ver)
	}

	gotBoot, err := admin.GetBootstrapToken(ctx)
	if err != nil || gotBoot == nil || *gotBoot != bootToken {
		t.Fatalf("GetBootstrapToken: err=%v tok=%v", err, gotBoot)
	}

	// 2. Disk & Bucket lifecycle
	disk, err := admin.AddDisk(ctx, fasts3.CreateDiskRequest{
		ID:   fasts3.StringPtr("go_disk"),
		Name: "Go Secondary Disk",
		Path: extraDiskPath,
	})
	if err != nil || disk.ID != "go_disk" {
		t.Fatalf("AddDisk: %v", err)
	}
	if _, err := admin.CreateBucket(ctx, fasts3.CreateBucketRequest{
		Name:     "go-e2e-bkt",
		IsPublic: false,
		DiskID:   fasts3.StringPtr("go_disk"),
		Quota:    50 * 1024 * 1024,
	}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}
	bkt, err := admin.GetBucketByName(ctx, "go-e2e-bkt")
	if err != nil || bkt == nil {
		t.Fatalf("GetBucketByName: %v", err)
	}
	if _, err := admin.UpdateBucket(ctx, bkt.ID, fasts3.UpdateBucketRequest{
		Name:     bkt.Name,
		IsPublic: false,
		DiskID:   bkt.DiskID,
		Quota:    60 * 1024 * 1024,
	}); err != nil {
		t.Fatalf("UpdateBucket: %v", err)
	}

	// 3. User, Permissions & API Key RBAC
	worker, err := admin.CreateUser(ctx, fasts3.CreateUserRequest{
		Username: "go_worker",
		Password: "Worker123!",
		IsAdmin:  fasts3.BoolPtr(false),
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := admin.UpdateUser(ctx, worker.ID, fasts3.UpdateUserRequest{
		Password: fasts3.StringPtr("Worker456!"),
	}); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if _, err := admin.SetPermission(ctx, fasts3.SetPermissionRequest{
		UserID:     worker.ID,
		BucketName: "go-e2e-bkt",
		CanRead:    true,
		CanWrite:   true,
		CanDelete:  true,
	}); err != nil {
		t.Fatalf("SetPermission: %v", err)
	}

	bPerms, err := admin.GetBucketPermissions(ctx, "go-e2e-bkt")
	if err != nil || len(bPerms) != 1 {
		t.Fatalf("GetBucketPermissions: err=%v len=%d", err, len(bPerms))
	}

	createdKey, err := admin.CreateAPIKey(ctx, fasts3.CreateAPIKeyRequest{
		Name:   "Go Worker Key",
		UserID: fasts3.Int64Ptr(worker.ID),
	})
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	workerCli, err := fasts3.NewClient(liveURL, fasts3.WithAPIKey(createdKey.APIKey))
	if err != nil {
		t.Fatal(err)
	}
	wVer, err := workerCli.VerifyAuth(ctx)
	if err != nil || wVer.IsAdmin || len(wVer.Permissions) != 1 {
		t.Fatalf("worker VerifyAuth: err=%v ver=%+v", err, wVer)
	}

	// 4. Uploads, Thumbnails, Downloads, Direct URLs
	if _, err := workerCli.UploadBytes(ctx, "go-e2e-bkt", "images/dot.png", tinyPNG, "image/png"); err != nil {
		t.Fatalf("UploadBytes: %v", err)
	}
	if _, err := workerCli.UploadJSONBytes(ctx, "go-e2e-bkt", "notes/hola mundo.txt", fasts3.StringPtr("hola mundo.txt"), []byte("Live Go SDK")); err != nil {
		t.Fatalf("UploadJSONBytes: %v", err)
	}

	syncSrc := filepath.Join(tempRoot, "sync_src")
	if err := os.MkdirAll(filepath.Join(syncSrc, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(syncSrc, "nested", "item.txt"), []byte("Go Synced"), 0o644); err != nil {
		t.Fatal(err)
	}
	uploaded, err := workerCli.UploadDirectory(ctx, "go-e2e-bkt", "batch", syncSrc)
	if err != nil || len(uploaded) != 1 {
		t.Fatalf("UploadDirectory: err=%v keys=%v", err, uploaded)
	}

	details, err := workerCli.ListFileDetails(ctx, "go-e2e-bkt", nil)
	if err != nil || len(details) != 3 {
		t.Fatalf("ListFileDetails: err=%v len=%d", err, len(details))
	}

	dl, err := workerCli.DownloadFile(ctx, "go-e2e-bkt", "notes/hola mundo.txt")
	if err != nil || string(dl.Data) != "Live Go SDK" {
		t.Fatalf("DownloadFile: err=%v dl=%+v", err, dl)
	}

	thumbOpts := &fasts3.ThumbnailOptions{Width: 64, Height: 64}
	thumb1, err := workerCli.GetThumbnail(ctx, "go-e2e-bkt", "images/dot.png", thumbOpts)
	if err != nil || len(thumb1.Data) == 0 {
		t.Fatalf("GetThumbnail: %v", err)
	}
	thumb2, err := workerCli.GetThumbnailAlt(ctx, "go-e2e-bkt", "images/dot.png", thumbOpts)
	if err != nil || len(thumb2.Data) == 0 {
		t.Fatalf("GetThumbnailAlt: %v", err)
	}

	// Verify direct query-authenticated URL works in raw HTTP GET
	directURL, err := workerCli.FileURL("go-e2e-bkt", "notes/hola mundo.txt", true)
	if err != nil {
		t.Fatal(err)
	}
	rawResp, err := http.Get(directURL)
	if err != nil {
		t.Fatalf("http.Get(directURL): %v", err)
	}
	rawBody, _ := io.ReadAll(rawResp.Body)
	_ = rawResp.Body.Close()
	if rawResp.StatusCode != http.StatusOK || string(rawBody) != "Live Go SDK" {
		t.Fatalf("directURL failed: status=%d body=%q", rawResp.StatusCode, string(rawBody))
	}

	pulledDir := filepath.Join(tempRoot, "pulled")
	pulled, err := workerCli.DownloadPrefixToDir(ctx, "go-e2e-bkt", "batch", pulledDir)
	if err != nil || len(pulled) != 1 {
		t.Fatalf("DownloadPrefixToDir: err=%v pulled=%v", err, pulled)
	}

	// 5. Cleanup & AutoVacuum
	delCount, err := workerCli.DeletePrefix(ctx, "go-e2e-bkt", "batch")
	if err != nil || delCount != 1 {
		t.Fatalf("DeletePrefix: err=%v count=%d", err, delCount)
	}
	if _, err := workerCli.DeleteFile(ctx, "go-e2e-bkt", "notes/hola mundo.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if _, err := workerCli.DeleteFile(ctx, "go-e2e-bkt", "images/dot.png"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}

	if _, err := admin.DeleteAPIKey(ctx, createdKey.ID); err != nil {
		t.Fatalf("DeleteAPIKey: %v", err)
	}
	if _, err := admin.DeleteBucketPermission(ctx, "go-e2e-bkt", worker.ID); err != nil {
		t.Fatalf("DeleteBucketPermission: %v", err)
	}
	if _, err := admin.DeleteUser(ctx, worker.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	vac, err := admin.TriggerVacuum(ctx)
	if err != nil || vac.BucketsReconciled < 1 {
		t.Fatalf("TriggerVacuum: err=%v vac=%+v", err, vac)
	}
	if _, err := admin.VacuumStatus(ctx); err != nil {
		t.Fatalf("VacuumStatus: %v", err)
	}

	if _, err := admin.DeleteBucketByName(ctx, "go-e2e-bkt"); err != nil {
		t.Fatalf("DeleteBucketByName: %v", err)
	}
	if _, err := admin.DeleteDisk(ctx, "go_disk"); err != nil {
		t.Fatalf("DeleteDisk: %v", err)
	}
}
