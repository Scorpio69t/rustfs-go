# Design Document: Missing S3 Features Implementation

## Architecture Overview

### Current Architecture

```
rustfs-go/
├── bucket/          # Bucket operations service
├── object/          # Object operations service
├── types/           # Common types
├── pkg/
│   └── credentials/ # Authentication
└── internal/
    ├── signer/      # Request signing
    ├── transport/   # HTTP transport
    └── core/        # Core executor
```

### Proposed Extensions

```
rustfs-go/
├── bucket/          # Extended with CORS, encryption, replication, notification
├── object/          # Extended with encryption, locking, ACL, compose, append
├── types/           # Add new types: ACL, ObjectLock, etc.
├── pkg/
│   ├── credentials/
│   ├── cors/        # NEW: CORS configuration
│   ├── sse/         # NEW: Server-side encryption
│   ├── objectlock/  # NEW: Object locking
│   ├── acl/         # NEW: Access control lists
│   ├── replication/ # NEW: Replication config (may exist)
│   ├── notification/# NEW: Event notification (may exist)
│   └── select/      # NEW: Select query
└── internal/
    ├── signer/      # Support SSE-C headers
    └── encryption/  # NEW: Encryption helpers
```

## Module Design

### 1. Server-Side Encryption (pkg/sse)

```go
package sse

// Configuration represents bucket encryption configuration
type Configuration struct {
    XMLName xml.Name `xml:"ServerSideEncryptionConfiguration"`
    Rules   []Rule   `xml:"Rule"`
}

type Rule struct {
    ApplySSEByDefault       DefaultEncryption `xml:"ApplyServerSideEncryptionByDefault"`
    BucketKeyEnabled        bool              `xml:"BucketKeyEnabled,omitempty"`
}

type DefaultEncryption struct {
    SSEAlgorithm   string `xml:"SSEAlgorithm"`           // AES256 or aws:kms
    KMSMasterKeyID string `xml:"KMSMasterKeyID,omitempty"`
}

// Options for object encryption
type SSEOption interface {
    ApplyToHeaders(h http.Header)
}

type SSES3 struct{}
func (s SSES3) ApplyToHeaders(h http.Header) {
    h.Set("x-amz-server-side-encryption", "AES256")
}

type SSEC struct {
    Key       []byte
    Algorithm string // AES256
}
func (s SSEC) ApplyToHeaders(h http.Header) {
    h.Set("x-amz-server-side-encryption-customer-algorithm", s.Algorithm)
    h.Set("x-amz-server-side-encryption-customer-key", base64.StdEncoding.EncodeToString(s.Key))
    md5sum := md5.Sum(s.Key)
    h.Set("x-amz-server-side-encryption-customer-key-md5", base64.StdEncoding.EncodeToString(md5sum[:]))
}

type SSEKMS struct {
    KeyID   string
    Context map[string]string
}
func (s SSEKMS) ApplyToHeaders(h http.Header) {
    h.Set("x-amz-server-side-encryption", "aws:kms")
    if s.KeyID != "" {
        h.Set("x-amz-server-side-encryption-aws-kms-key-id", s.KeyID)
    }
    if len(s.Context) > 0 {
        ctx, _ := json.Marshal(s.Context)
        h.Set("x-amz-server-side-encryption-context", base64.StdEncoding.EncodeToString(ctx))
    }
}
```

### 2. CORS Configuration (pkg/cors)

```go
package cors

type Config struct {
    XMLName   xml.Name   `xml:"CORSConfiguration"`
    CORSRules []CORSRule `xml:"CORSRule"`
}

type CORSRule struct {
    AllowedOrigins []string `xml:"AllowedOrigin"`
    AllowedMethods []string `xml:"AllowedMethod"`
    AllowedHeaders []string `xml:"AllowedHeader,omitempty"`
    ExposeHeaders  []string `xml:"ExposeHeader,omitempty"`
    MaxAgeSeconds  int      `xml:"MaxAgeSeconds,omitempty"`
}

// Builder pattern for easy construction
type ConfigBuilder struct {
    rules []CORSRule
}

func NewConfigBuilder() *ConfigBuilder {
    return &ConfigBuilder{}
}

func (b *ConfigBuilder) AddRule(origins, methods []string) *CORSRule {
    rule := CORSRule{
        AllowedOrigins: origins,
        AllowedMethods: methods,
    }
    b.rules = append(b.rules, rule)
    return &b.rules[len(b.rules)-1]
}

func (r *CORSRule) WithHeaders(headers []string) *CORSRule {
    r.AllowedHeaders = headers
    return r
}

func (r *CORSRule) WithExposeHeaders(headers []string) *CORSRule {
    r.ExposeHeaders = headers
    return r
}

func (r *CORSRule) WithMaxAge(seconds int) *CORSRule {
    r.MaxAgeSeconds = seconds
    return r
}

func (b *ConfigBuilder) Build() Config {
    return Config{CORSRules: b.rules}
}
```

### 3. Object Locking (pkg/objectlock)

```go
package objectlock

type Config struct {
    XMLName           xml.Name `xml:"ObjectLockConfiguration"`
    ObjectLockEnabled string   `xml:"ObjectLockEnabled"` // "Enabled"
    Rule              *Rule    `xml:"Rule,omitempty"`
}

type Rule struct {
    DefaultRetention DefaultRetention `xml:"DefaultRetention"`
}

type DefaultRetention struct {
    Mode  RetentionMode `xml:"Mode"`  // GOVERNANCE or COMPLIANCE
    Days  int           `xml:"Days,omitempty"`
    Years int           `xml:"Years,omitempty"`
}

type RetentionMode string

const (
    Governance RetentionMode = "GOVERNANCE"
    Compliance RetentionMode = "COMPLIANCE"
)

type LegalHoldStatus string

const (
    LegalHoldOn  LegalHoldStatus = "ON"
    LegalHoldOff LegalHoldStatus = "OFF"
)

type Retention struct {
    XMLName         xml.Name      `xml:"Retention"`
    Mode            RetentionMode `xml:"Mode"`
    RetainUntilDate time.Time     `xml:"RetainUntilDate"`
}

type LegalHold struct {
    XMLName xml.Name        `xml:"LegalHold"`
    Status  LegalHoldStatus `xml:"Status"`
}
```

### 4. ACL (pkg/acl)

```go
package acl

type AccessControlPolicy struct {
    XMLName xml.Name `xml:"AccessControlPolicy"`
    Owner   Owner    `xml:"Owner"`
    Grants  []Grant  `xml:"AccessControlList>Grant"`
}

type Owner struct {
    ID          string `xml:"ID"`
    DisplayName string `xml:"DisplayName,omitempty"`
}

type Grant struct {
    Grantee    Grantee    `xml:"Grantee"`
    Permission Permission `xml:"Permission"`
}

type Grantee struct {
    XMLName      xml.Name `xml:"Grantee"`
    Type         string   `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
    ID           string   `xml:"ID,omitempty"`
    DisplayName  string   `xml:"DisplayName,omitempty"`
    EmailAddress string   `xml:"EmailAddress,omitempty"`
    URI          string   `xml:"URI,omitempty"`
}

type Permission string

const (
    PermissionFullControl Permission = "FULL_CONTROL"
    PermissionWrite       Permission = "WRITE"
    PermissionWriteACP    Permission = "WRITE_ACP"
    PermissionRead        Permission = "READ"
    PermissionReadACP     Permission = "READ_ACP"
)

// Canned ACLs
type CannedACL string

const (
    ACLPrivate                CannedACL = "private"
    ACLPublicRead             CannedACL = "public-read"
    ACLPublicReadWrite        CannedACL = "public-read-write"
    ACLAuthenticatedRead      CannedACL = "authenticated-read"
    ACLBucketOwnerRead        CannedACL = "bucket-owner-read"
    ACLBucketOwnerFullControl CannedACL = "bucket-owner-full-control"
)
```

## API Design Principles

### 1. Consistency with Existing API

**Option Functions Pattern**:
```go
// Existing pattern
object.Put(ctx, bucket, name, reader, size,
    object.WithContentType("text/plain"),
    object.WithUserTags(tags),
)

// New encryption options
object.Put(ctx, bucket, name, reader, size,
    object.WithSSES3(),  // Server-side encryption
)

object.Put(ctx, bucket, name, reader, size,
    object.WithSSEC(encryptionKey),  // Customer-provided key
)
```

**Service-Based Methods**:
```go
// Existing pattern
bucketSvc := client.Bucket()
bucketSvc.SetPolicy(ctx, bucket, policy)

// New methods
bucketSvc.SetCORS(ctx, bucket, corsConfig)
bucketSvc.SetEncryption(ctx, bucket, encConfig)
```

### 2. Type Safety

Use strongly-typed enums and structs:
```go
// Good: Type-safe
mode := objectlock.Governance
status := objectlock.LegalHoldOn

// Avoid: String constants
mode := "GOVERNANCE"  // easy to mistype
```

### 3. Builder Patterns for Complex Types

```go
// CORS configuration
corsConfig := cors.NewConfigBuilder().
    AddRule([]string{"https://example.com"}, []string{"GET", "PUT"}).
        WithHeaders([]string{"*"}).
        WithMaxAge(3600).
    AddRule([]string{"*"}, []string{"GET"}).
    Build()

bucketSvc.SetCORS(ctx, bucket, corsConfig)
```

### 4. Error Handling

Follow existing error patterns:
```go
// Standard error returns
config, err := bucketSvc.GetCORS(ctx, bucket)
if err != nil {
    // Check specific error types
    if errors.Is(err, ErrNoCORSConfiguration) {
        // Handle missing config
    }
    return err
}
```

## Implementation Strategy

### Phase 1: Security Features (Week 1-2)

**Week 1: SSE Implementation**
1. Create `pkg/sse` package
2. Implement SSE-S3 and SSE-C
3. Add encryption options to `object.PutOption`
4. Add bucket encryption to `bucket` service
5. Unit tests

**Week 2: Object Locking**
1. Create `pkg/objectlock` package
2. Implement bucket lock config
3. Implement object retention
4. Implement legal hold
5. Unit tests

### Phase 2: Configuration (Week 3)

1. Create `pkg/cors` package
2. Implement CORS in bucket service
3. Create `pkg/acl` package
4. Implement ACL for objects and buckets
5. Bucket tagging (if needed)
6. Unit tests

### Phase 3: Advanced Features (Week 4-5)

**Week 4: Replication & Notification**
1. Implement/extend replication config
2. Implement event notification
3. Unit tests

**Week 5: Object Operations**
1. Implement object composition
2. Implement object append
3. Unit tests

### Phase 4: Query & Restore (Week 6)

1. Create `pkg/select` package
2. Implement Select query
3. Implement object restore
4. Post Policy implementation
5. Unit tests

### Phase 5: Examples & Documentation (Week 7-8)

1. Create 20+ examples
2. Update documentation
3. Integration tests
4. Performance tests

## Testing Strategy

### Unit Tests

```go
func TestSSES3Put(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify SSE header
        if r.Header.Get("x-amz-server-side-encryption") != "AES256" {
            t.Error("Missing SSE header")
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Test
    client, _ := rustfs.New(server.URL, &rustfs.Options{})
    svc := client.Object()

    _, err := svc.Put(ctx, "bucket", "object",
        strings.NewReader("data"), 4,
        object.WithSSES3(),
    )

    if err != nil {
        t.Fatal(err)
    }
}
```

### Integration Tests

```go
func TestCORSIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client := setupTestClient(t)
    bucketSvc := client.Bucket()

    // Create bucket
    bucketSvc.Create(ctx, testBucket)
    defer bucketSvc.Delete(ctx, testBucket)

    // Set CORS
    config := cors.NewConfigBuilder().
        AddRule([]string{"*"}, []string{"GET"}).
        Build()

    err := bucketSvc.SetCORS(ctx, testBucket, config)
    assert.NoError(t, err)

    // Get and verify
    retrieved, err := bucketSvc.GetCORS(ctx, testBucket)
    assert.NoError(t, err)
    assert.Equal(t, config, retrieved)
}
```

## Migration from Old SDK

### Mapping Table

| Old API | New API | Notes |
|---------|---------|-------|
| `client.SetBucketEncryption()` | `bucket.SetEncryption()` | Same functionality |
| `client.GetBucketEncryption()` | `bucket.GetEncryption()` | Returns typed config |
| `client.SetBucketCors()` | `bucket.SetCORS()` | Builder pattern available |
| `client.PutObjectWithSSE()` | `object.Put(..., object.WithSSES3())` | Option function |
| `client.PutObjectRetention()` | `object.SetRetention()` | Clearer naming |
| `client.ComposeObject()` | `object.Compose()` | Simplified API |

### Example Migration

**Old SDK**:
```go
// Old way
opts := minio.PutObjectOptions{}
opts.ServerSideEncryption = encrypt.NewSSE()
client.PutObject(ctx, bucket, object, reader, size, opts)
```

**New SDK**:
```go
// New way
svc := client.Object()
svc.Put(ctx, bucket, object, reader, size,
    object.WithSSES3(),
)
```

## Performance Considerations

1. **Encryption Overhead**: SSE-C requires client-side MD5 calculation
2. **Select Queries**: Stream results to avoid memory issues
3. **Replication**: Async operations, non-blocking
4. **Object Composition**: May involve multiple API calls

## Security Considerations

1. **Key Management**: SSE-C keys never sent to server unencrypted
2. **TLS Required**: Enforce HTTPS for encryption operations
3. **Validation**: Validate all XML input/output
4. **ACL Checks**: Verify permissions before operations

## Future Enhancements

1. **Batch Operations**: Batch ACL/tagging operations
2. **Async Replication**: Background replication monitoring
3. **Query Optimization**: Query result caching
4. **Metrics**: Operation metrics and monitoring
