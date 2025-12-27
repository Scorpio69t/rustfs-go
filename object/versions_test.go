// Package object object/versions_test.go
package object

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("versions") != "true" {
			t.Fatalf("expected versions query flag")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`
<ListVersionsResult>
  <Name>demo</Name>
  <Prefix></Prefix>
  <KeyMarker></KeyMarker>
  <VersionIdMarker></VersionIdMarker>
  <IsTruncated>false</IsTruncated>
  <Version>
    <Key>foo.txt</Key>
    <VersionId>v1</VersionId>
    <IsLatest>true</IsLatest>
    <LastModified>2024-01-01T00:00:00.000Z</LastModified>
    <ETag>"etag1"</ETag>
    <Size>10</Size>
    <StorageClass>STANDARD</StorageClass>
  </Version>
  <DeleteMarker>
    <Key>bar.txt</Key>
    <VersionId>v2</VersionId>
    <IsLatest>false</IsLatest>
    <LastModified>2024-01-02T00:00:00.000Z</LastModified>
  </DeleteMarker>
</ListVersionsResult>`))
	}))
	defer server.Close()

	svc := createTestService(t, server)

	var got []string
	for info := range svc.ListVersions(context.Background(), "demo", WithListRecursive(true)) {
		if info.Err != nil {
			t.Fatalf("ListVersions returned error: %v", info.Err)
		}
		got = append(got, info.Key+info.VersionID)
		if info.Key == "foo.txt" && (!info.IsLatest || info.IsDeleteMarker) {
			t.Fatalf("expected foo.txt latest version flag")
		}
		if info.Key == "bar.txt" && !info.IsDeleteMarker {
			t.Fatalf("expected delete marker for bar.txt")
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}
