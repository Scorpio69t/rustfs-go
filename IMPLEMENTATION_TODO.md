# RustFS Go SDK 重构实施计划

## 📋 总体进度追踪

| 阶段 | 任务数 | 完成 | 进度 |
|------|--------|------|------|
| 第一阶段：基础架构 | 12 | 0 | 0% |
| 第二阶段：核心模块 | 15 | 0 | 0% |
| 第三阶段：Bucket 模块 | 14 | 0 | 0% |
| 第四阶段：Object 模块 | 18 | 0 | 0% |
| 第五阶段：兼容层和测试 | 10 | 0 | 0% |
| **总计** | **69** | **0** | **0%** |

---

## 🚀 第一阶段：基础架构搭建（预计 5 天）

### 任务 1.1：创建目录结构
**状态**: ⬜ 未开始  
**预计时间**: 0.5 天

#### 实施步骤

```bash
# 1. 创建主要模块目录
mkdir -p bucket/config bucket/policy
mkdir -p object/upload object/download object/manage object/presign

# 2. 创建内部实现目录
mkdir -p internal/core
mkdir -p internal/signer
mkdir -p internal/transport
mkdir -p internal/cache
mkdir -p internal/xml

# 3. 创建公共目录
mkdir -p errors
mkdir -p types

# 4. 创建文档和示例目录
mkdir -p docs
mkdir -p examples/basic/upload
mkdir -p examples/basic/download
mkdir -p examples/basic/bucket
mkdir -p examples/advanced/multipart
mkdir -p examples/advanced/presign
```

#### 验证清单
- [ ] 所有目录已创建
- [ ] 目录结构符合设计方案
- [ ] `.gitkeep` 文件添加到空目录（可选）

---

### 任务 1.2：创建类型定义包 `types/`
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 1.2.1 创建 `types/common.go`

```go
// types/common.go
package types

import (
    "net/http"
    "time"
)

// Owner 对象所有者信息
type Owner struct {
    DisplayName string `json:"displayName,omitempty"`
    ID          string `json:"id,omitempty"`
}

// Grant ACL 授权
type Grant struct {
    Grantee    Grantee
    Permission string
}

// Grantee 授权对象
type Grantee struct {
    Type        string
    ID          string
    DisplayName string
    URI         string
}

// RestoreInfo 归档恢复信息
type RestoreInfo struct {
    OngoingRestore bool
    ExpiryTime     time.Time
}

// ChecksumType 校验和类型
type ChecksumType int

const (
    ChecksumNone ChecksumType = iota
    ChecksumCRC32
    ChecksumCRC32C
    ChecksumSHA1
    ChecksumSHA256
    ChecksumCRC64NVME
)

// String 返回校验和类型字符串
func (c ChecksumType) String() string {
    switch c {
    case ChecksumCRC32:
        return "CRC32"
    case ChecksumCRC32C:
        return "CRC32C"
    case ChecksumSHA1:
        return "SHA1"
    case ChecksumSHA256:
        return "SHA256"
    case ChecksumCRC64NVME:
        return "CRC64NVME"
    default:
        return ""
    }
}

// RetentionMode 保留模式
type RetentionMode string

const (
    RetentionGovernance RetentionMode = "GOVERNANCE"
    RetentionCompliance RetentionMode = "COMPLIANCE"
)

// IsValid 验证保留模式是否有效
func (r RetentionMode) IsValid() bool {
    return r == RetentionGovernance || r == RetentionCompliance
}

// LegalHoldStatus 法律保留状态
type LegalHoldStatus string

const (
    LegalHoldOn  LegalHoldStatus = "ON"
    LegalHoldOff LegalHoldStatus = "OFF"
)

// IsValid 验证法律保留状态是否有效
func (l LegalHoldStatus) IsValid() bool {
    return l == LegalHoldOn || l == LegalHoldOff
}

// ReplicationStatus 复制状态
type ReplicationStatus string

const (
    ReplicationPending  ReplicationStatus = "PENDING"
    ReplicationComplete ReplicationStatus = "COMPLETED"
    ReplicationFailed   ReplicationStatus = "FAILED"
    ReplicationReplica  ReplicationStatus = "REPLICA"
)

// StringMap 自定义字符串映射（用于 XML 解析）
type StringMap map[string]string

// URLMap URL 编码的映射
type URLMap map[string]string
```

#### 1.2.2 创建 `types/bucket.go`

```go
// types/bucket.go
package types

import "time"

// BucketInfo 桶信息
type BucketInfo struct {
    // 桶名称
    Name string `json:"name"`
    // 创建时间
    CreationDate time.Time `json:"creationDate"`
    // 桶所在区域
    Region string `json:"region,omitempty"`
}

// BucketLookupType 桶查找类型
type BucketLookupType int

const (
    // BucketLookupAuto 自动检测
    BucketLookupAuto BucketLookupType = iota
    // BucketLookupDNS DNS 风格
    BucketLookupDNS
    // BucketLookupPath 路径风格
    BucketLookupPath
)

// VersioningConfig 版本控制配置
type VersioningConfig struct {
    Status    string // Enabled, Suspended
    MFADelete string // Enabled, Disabled
}

// IsEnabled 检查版本控制是否启用
func (v VersioningConfig) IsEnabled() bool {
    return v.Status == "Enabled"
}

// IsSuspended 检查版本控制是否暂停
func (v VersioningConfig) IsSuspended() bool {
    return v.Status == "Suspended"
}
```

#### 1.2.3 创建 `types/object.go`

```go
// types/object.go
package types

import (
    "net/http"
    "time"
)

// ObjectInfo 对象元数据信息
type ObjectInfo struct {
    // 基本信息
    Key          string    `json:"name"`
    Size         int64     `json:"size"`
    ETag         string    `json:"etag"`
    ContentType  string    `json:"contentType"`
    LastModified time.Time `json:"lastModified"`
    Expires      time.Time `json:"expires,omitempty"`

    // 所有者
    Owner Owner `json:"owner,omitempty"`

    // 存储类
    StorageClass string `json:"storageClass,omitempty"`

    // 版本信息
    VersionID      string `json:"versionId,omitempty"`
    IsLatest       bool   `json:"isLatest,omitempty"`
    IsDeleteMarker bool   `json:"isDeleteMarker,omitempty"`

    // 复制状态
    ReplicationStatus string `json:"replicationStatus,omitempty"`

    // 元数据
    Metadata     http.Header `json:"metadata,omitempty"`
    UserMetadata StringMap   `json:"userMetadata,omitempty"`
    UserTags     URLMap      `json:"userTags,omitempty"`
    UserTagCount int         `json:"userTagCount,omitempty"`

    // 生命周期
    Expiration       time.Time `json:"expiration,omitempty"`
    ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

    // 恢复信息
    Restore *RestoreInfo `json:"restore,omitempty"`

    // 校验和
    ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
    ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
    ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
    ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
    ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
    ChecksumMode      string `json:"checksumMode,omitempty"`

    // ACL
    Grant []Grant `json:"grant,omitempty"`

    // 版本数量
    NumVersions int `json:"numVersions,omitempty"`

    // 内部信息（EC 编码）
    Internal *struct {
        K int
        M int
    } `json:"internal,omitempty"`

    // 错误（用于列表操作）
    Err error `json:"-"`
}

// ObjectToDelete 待删除对象
type ObjectToDelete struct {
    Key       string
    VersionID string
}

// DeletedObject 已删除对象结果
type DeletedObject struct {
    Key                   string
    VersionID             string
    DeleteMarker          bool
    DeleteMarkerVersionID string
}

// DeleteError 删除错误
type DeleteError struct {
    Key       string
    VersionID string
    Code      string
    Message   string
}
```

#### 1.2.4 创建 `types/upload.go`

```go
// types/upload.go
package types

import "time"

// UploadInfo 上传结果信息
type UploadInfo struct {
    // 桶名称
    Bucket string `json:"bucket"`
    // 对象键
    Key string `json:"key"`
    // ETag
    ETag string `json:"etag"`
    // 大小
    Size int64 `json:"size"`
    // 最后修改时间
    LastModified time.Time `json:"lastModified"`
    // 位置
    Location string `json:"location,omitempty"`
    // 版本 ID
    VersionID string `json:"versionId,omitempty"`

    // 生命周期过期信息
    Expiration       time.Time `json:"expiration,omitempty"`
    ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

    // 校验和
    ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
    ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
    ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
    ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
    ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
    ChecksumMode      string `json:"checksumMode,omitempty"`
}

// MultipartInfo 分片上传信息
type MultipartInfo struct {
    // 上传 ID
    UploadID string `json:"uploadId"`
    // 对象键
    Key string `json:"key"`
    // 发起时间
    Initiated time.Time `json:"initiated"`
    // 发起者
    Initiator struct {
        ID          string
        DisplayName string
    } `json:"initiator,omitempty"`
    // 所有者
    Owner Owner `json:"owner,omitempty"`
    // 存储类
    StorageClass string `json:"storageClass,omitempty"`
    // 大小（聚合）
    Size int64 `json:"size,omitempty"`
    // 错误
    Err error `json:"-"`
}

// PartInfo 分片信息
type PartInfo struct {
    // 分片号
    PartNumber int `json:"partNumber"`
    // ETag
    ETag string `json:"etag"`
    // 大小
    Size int64 `json:"size"`
    // 最后修改时间
    LastModified time.Time `json:"lastModified"`

    // 校验和
    ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
    ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
    ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
    ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
    ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
}

// CompletePart 完成分片信息
type CompletePart struct {
    PartNumber        int
    ETag              string
    ChecksumCRC32     string
    ChecksumCRC32C    string
    ChecksumSHA1      string
    ChecksumSHA256    string
    ChecksumCRC64NVME string
}
```

#### 验证清单
- [ ] `types/common.go` 已创建
- [ ] `types/bucket.go` 已创建
- [ ] `types/object.go` 已创建
- [ ] `types/upload.go` 已创建
- [ ] 所有类型编译通过
- [ ] GoDoc 注释完整

---

### 任务 1.3：创建错误定义包 `errors/`
**状态**: ⬜ 未开始  
**预计时间**: 0.5 天

#### 1.3.1 创建 `errors/codes.go`

```go
// errors/codes.go
package errors

// S3 标准错误码
const (
    // 桶相关
    ErrCodeNoSuchBucket           = "NoSuchBucket"
    ErrCodeBucketAlreadyExists    = "BucketAlreadyExists"
    ErrCodeBucketAlreadyOwnedByYou = "BucketAlreadyOwnedByYou"
    ErrCodeBucketNotEmpty         = "BucketNotEmpty"
    ErrCodeInvalidBucketName      = "InvalidBucketName"

    // 对象相关
    ErrCodeNoSuchKey            = "NoSuchKey"
    ErrCodeInvalidObjectName    = "XMinioInvalidObjectName"
    ErrCodeNoSuchUpload         = "NoSuchUpload"
    ErrCodeNoSuchVersion        = "NoSuchVersion"
    ErrCodeInvalidPart          = "InvalidPart"
    ErrCodeInvalidPartOrder     = "InvalidPartOrder"
    ErrCodeEntityTooLarge       = "EntityTooLarge"
    ErrCodeEntityTooSmall       = "EntityTooSmall"

    // 访问控制
    ErrCodeAccessDenied         = "AccessDenied"
    ErrCodeAccountProblem       = "AccountProblem"
    ErrCodeInvalidAccessKeyId   = "InvalidAccessKeyId"
    ErrCodeSignatureDoesNotMatch = "SignatureDoesNotMatch"

    // 请求相关
    ErrCodeInvalidArgument      = "InvalidArgument"
    ErrCodeInvalidRequest       = "InvalidRequest"
    ErrCodeMalformedXML         = "MalformedXML"
    ErrCodeMissingContentLength = "MissingContentLength"
    ErrCodeMethodNotAllowed     = "MethodNotAllowed"

    // 区域相关
    ErrCodeInvalidRegion                = "InvalidRegion"
    ErrCodeAuthorizationHeaderMalformed = "AuthorizationHeaderMalformed"

    // 服务器
    ErrCodeInternalError    = "InternalError"
    ErrCodeServiceUnavailable = "ServiceUnavailable"
    ErrCodeSlowDown         = "SlowDown"
    ErrCodeNotImplemented   = "NotImplemented"

    // 条件请求
    ErrCodePreconditionFailed = "PreconditionFailed"
    ErrCodeNotModified        = "NotModified"

    // 复制
    ErrCodeInvalidCopySource = "InvalidCopySource"
)

// HTTP 状态码到错误码的映射
var httpStatusToCode = map[int]string{
    301: "MovedPermanently",
    400: ErrCodeInvalidArgument,
    403: ErrCodeAccessDenied,
    404: ErrCodeNoSuchKey,
    405: ErrCodeMethodNotAllowed,
    409: "Conflict",
    411: ErrCodeMissingContentLength,
    412: ErrCodePreconditionFailed,
    416: "InvalidRange",
    500: ErrCodeInternalError,
    501: ErrCodeNotImplemented,
    503: ErrCodeServiceUnavailable,
}
```

#### 1.3.2 创建 `errors/errors.go`

```go
// errors/errors.go
package errors

import (
    "encoding/xml"
    "fmt"
    "io"
    "net/http"
)

// Error RustFS 错误接口
type Error interface {
    error
    Code() string
    Message() string
    StatusCode() int
    RequestID() string
    Resource() string
}

// APIError S3 API 错误
type APIError struct {
    XMLName    xml.Name `xml:"Error"`
    code       string   `xml:"Code"`
    message    string   `xml:"Message"`
    resource   string   `xml:"Resource"`
    requestID  string   `xml:"RequestId"`
    hostID     string   `xml:"HostId"`
    statusCode int      `xml:"-"`
    region     string   `xml:"Region"`
}

// NewAPIError 创建新的 API 错误
func NewAPIError(code, message string, statusCode int) *APIError {
    return &APIError{
        code:       code,
        message:    message,
        statusCode: statusCode,
    }
}

// Error 实现 error 接口
func (e *APIError) Error() string {
    if e.requestID != "" {
        return fmt.Sprintf("%s: %s (RequestID: %s)", e.code, e.message, e.requestID)
    }
    return fmt.Sprintf("%s: %s", e.code, e.message)
}

// Code 返回错误码
func (e *APIError) Code() string { return e.code }

// Message 返回错误信息
func (e *APIError) Message() string { return e.message }

// StatusCode 返回 HTTP 状态码
func (e *APIError) StatusCode() int { return e.statusCode }

// RequestID 返回请求 ID
func (e *APIError) RequestID() string { return e.requestID }

// Resource 返回资源路径
func (e *APIError) Resource() string { return e.resource }

// Region 返回区域
func (e *APIError) Region() string { return e.region }

// HostID 返回主机 ID
func (e *APIError) HostID() string { return e.hostID }

// WithRequestID 设置请求 ID
func (e *APIError) WithRequestID(id string) *APIError {
    e.requestID = id
    return e
}

// WithResource 设置资源
func (e *APIError) WithResource(resource string) *APIError {
    e.resource = resource
    return e
}

// WithRegion 设置区域
func (e *APIError) WithRegion(region string) *APIError {
    e.region = region
    return e
}

// ParseErrorResponse 从 HTTP 响应解析错误
func ParseErrorResponse(resp *http.Response, bucketName, objectName string) error {
    if resp == nil {
        return NewAPIError(ErrCodeInternalError, "empty response", 500)
    }

    // 读取响应体
    body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 最大 1MB
    if err != nil {
        return NewAPIError(ErrCodeInternalError, "failed to read response body", resp.StatusCode)
    }

    // 尝试解析 XML 错误响应
    apiErr := &APIError{statusCode: resp.StatusCode}
    if len(body) > 0 {
        if xmlErr := xml.Unmarshal(body, apiErr); xmlErr == nil {
            apiErr.statusCode = resp.StatusCode
            return apiErr
        }
    }

    // 使用状态码生成错误
    code := httpStatusToCode[resp.StatusCode]
    if code == "" {
        code = ErrCodeInternalError
    }

    return &APIError{
        code:       code,
        message:    http.StatusText(resp.StatusCode),
        statusCode: resp.StatusCode,
        requestID:  resp.Header.Get("x-amz-request-id"),
        hostID:     resp.Header.Get("x-amz-id-2"),
        resource:   "/" + bucketName + "/" + objectName,
    }
}
```

#### 1.3.3 创建 `errors/check.go`

```go
// errors/check.go
package errors

import "errors"

// IsNotFound 检查是否为未找到错误
func IsNotFound(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeNoSuchBucket || 
               apiErr.Code() == ErrCodeNoSuchKey ||
               apiErr.Code() == ErrCodeNoSuchUpload
    }
    return false
}

// IsBucketNotFound 检查桶是否不存在
func IsBucketNotFound(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeNoSuchBucket
    }
    return false
}

// IsObjectNotFound 检查对象是否不存在
func IsObjectNotFound(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeNoSuchKey
    }
    return false
}

// IsAccessDenied 检查是否为访问拒绝错误
func IsAccessDenied(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeAccessDenied
    }
    return false
}

// IsBucketExists 检查桶是否已存在
func IsBucketExists(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeBucketAlreadyExists ||
               apiErr.Code() == ErrCodeBucketAlreadyOwnedByYou
    }
    return false
}

// IsBucketNotEmpty 检查桶是否非空
func IsBucketNotEmpty(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeBucketNotEmpty
    }
    return false
}

// IsInvalidArgument 检查是否为无效参数错误
func IsInvalidArgument(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeInvalidArgument
    }
    return false
}

// IsServiceUnavailable 检查服务是否不可用
func IsServiceUnavailable(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code() == ErrCodeServiceUnavailable ||
               apiErr.Code() == ErrCodeSlowDown
    }
    return false
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Code() {
        case ErrCodeServiceUnavailable,
             ErrCodeSlowDown,
             ErrCodeInternalError,
             "RequestTimeout",
             "RequestTimeTooSkewed":
            return true
        }
        // 5xx 错误通常可重试
        if apiErr.StatusCode() >= 500 {
            return true
        }
    }
    return false
}

// ToAPIError 将错误转换为 APIError
func ToAPIError(err error) *APIError {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        return apiErr
    }
    return nil
}
```

#### 验证清单
- [ ] `errors/codes.go` 已创建
- [ ] `errors/errors.go` 已创建
- [ ] `errors/check.go` 已创建
- [ ] 编译通过
- [ ] 单元测试通过

---

### 任务 1.4：创建内部核心包 `internal/core/`
**状态**: ⬜ 未开始  
**预计时间**: 1.5 天

#### 1.4.1 创建 `internal/core/request.go`

```go
// internal/core/request.go
package core

import (
    "context"
    "io"
    "net/http"
    "net/url"
)

// RequestMetadata 请求元数据
type RequestMetadata struct {
    // 桶和对象
    BucketName string
    ObjectName string

    // 查询参数
    QueryValues url.Values

    // 请求头
    CustomHeader http.Header

    // 请求体
    ContentBody   io.Reader
    ContentLength int64

    // 内容校验
    ContentMD5Base64 string
    ContentSHA256Hex string

    // 签名选项
    StreamSHA256 bool
    PresignURL   bool
    Expires      int64

    // 预签名额外头
    ExtraPresignHeader http.Header

    // 位置
    BucketLocation string

    // Trailer (用于流式签名)
    Trailer http.Header
    AddCRC  bool

    // 特殊处理
    Expect200OKWithError bool
}

// Request 封装的 HTTP 请求
type Request struct {
    ctx      context.Context
    method   string
    metadata RequestMetadata
}

// NewRequest 创建新请求
func NewRequest(ctx context.Context, method string, metadata RequestMetadata) *Request {
    return &Request{
        ctx:      ctx,
        method:   method,
        metadata: metadata,
    }
}

// Context 返回请求上下文
func (r *Request) Context() context.Context {
    return r.ctx
}

// Method 返回 HTTP 方法
func (r *Request) Method() string {
    return r.method
}

// Metadata 返回请求元数据
func (r *Request) Metadata() RequestMetadata {
    return r.metadata
}
```

#### 1.4.2 创建 `internal/core/executor.go`

```go
// internal/core/executor.go
package core

import (
    "context"
    "io"
    "net/http"
    "net/url"
    "time"

    "github.com/Scorpio69t/rustfs-go/errors"
    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

// Executor 请求执行器
type Executor struct {
    // HTTP 客户端
    httpClient *http.Client

    // 端点
    endpointURL *url.URL

    // 凭证
    credentials *credentials.Credentials

    // 区域
    region string

    // 是否使用 HTTPS
    secure bool

    // 签名类型
    signerType credentials.SignatureType

    // 桶查找方式
    bucketLookup int

    // 最大重试次数
    maxRetries int

    // 位置缓存
    locationCache LocationCache

    // 调试选项
    traceEnabled bool
    traceOutput  io.Writer
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
    HTTPClient    *http.Client
    EndpointURL   *url.URL
    Credentials   *credentials.Credentials
    Region        string
    Secure        bool
    BucketLookup  int
    MaxRetries    int
    LocationCache LocationCache
}

// NewExecutor 创建新的执行器
func NewExecutor(config ExecutorConfig) *Executor {
    maxRetries := config.MaxRetries
    if maxRetries <= 0 {
        maxRetries = 10
    }

    return &Executor{
        httpClient:    config.HTTPClient,
        endpointURL:   config.EndpointURL,
        credentials:   config.Credentials,
        region:        config.Region,
        secure:        config.Secure,
        bucketLookup:  config.BucketLookup,
        maxRetries:    maxRetries,
        locationCache: config.LocationCache,
    }
}

// Execute 执行请求
func (e *Executor) Execute(ctx context.Context, req *Request) (*http.Response, error) {
    var (
        resp *http.Response
        err  error
    )

    // 重试循环
    for attempt := 0; attempt < e.maxRetries; attempt++ {
        // 检查上下文
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }

        // 构建 HTTP 请求
        httpReq, err := e.buildHTTPRequest(ctx, req)
        if err != nil {
            return nil, err
        }

        // 执行请求
        resp, err = e.httpClient.Do(httpReq)
        if err != nil {
            if e.shouldRetry(err, attempt) {
                e.waitForRetry(ctx, attempt)
                continue
            }
            return nil, err
        }

        // 检查响应
        if e.isSuccessStatus(resp.StatusCode, req.metadata.Expect200OKWithError) {
            return resp, nil
        }

        // 检查是否需要重试
        if e.shouldRetryResponse(resp, attempt) {
            closeResponse(resp)
            e.waitForRetry(ctx, attempt)
            continue
        }

        // 返回错误响应
        return resp, nil
    }

    if err != nil {
        return nil, err
    }

    return resp, nil
}

// buildHTTPRequest 构建 HTTP 请求
func (e *Executor) buildHTTPRequest(ctx context.Context, req *Request) (*http.Request, error) {
    meta := req.Metadata()

    // 获取桶位置
    location := meta.BucketLocation
    if location == "" && meta.BucketName != "" {
        location = e.getBucketLocation(ctx, meta.BucketName)
    }
    if location == "" {
        location = e.region
    }

    // 构建 URL
    targetURL, err := e.makeTargetURL(meta.BucketName, meta.ObjectName, location, meta.QueryValues)
    if err != nil {
        return nil, err
    }

    // 创建请求
    httpReq, err := http.NewRequestWithContext(ctx, req.Method(), targetURL.String(), meta.ContentBody)
    if err != nil {
        return nil, err
    }

    // 设置头部
    for k, v := range meta.CustomHeader {
        httpReq.Header[k] = v
    }

    // 设置 Content-Length
    httpReq.ContentLength = meta.ContentLength

    // 签名请求
    if err := e.signRequest(httpReq, meta, location); err != nil {
        return nil, err
    }

    return httpReq, nil
}

// makeTargetURL 构建目标 URL
func (e *Executor) makeTargetURL(bucketName, objectName, location string, queryValues url.Values) (*url.URL, error) {
    // TODO: 实现 URL 构建逻辑
    // 根据 bucketLookup 决定使用路径风格还是虚拟主机风格
    return nil, nil
}

// signRequest 签名请求
func (e *Executor) signRequest(req *http.Request, meta RequestMetadata, location string) error {
    // TODO: 实现签名逻辑
    return nil
}

// getBucketLocation 获取桶位置
func (e *Executor) getBucketLocation(ctx context.Context, bucketName string) string {
    if e.locationCache != nil {
        if loc, ok := e.locationCache.Get(bucketName); ok {
            return loc
        }
    }
    return e.region
}

// shouldRetry 判断是否应该重试
func (e *Executor) shouldRetry(err error, attempt int) bool {
    if attempt >= e.maxRetries-1 {
        return false
    }
    // TODO: 检查网络错误等
    return false
}

// shouldRetryResponse 判断响应是否应该重试
func (e *Executor) shouldRetryResponse(resp *http.Response, attempt int) bool {
    if attempt >= e.maxRetries-1 {
        return false
    }
    // 5xx 错误可重试
    if resp.StatusCode >= 500 {
        return true
    }
    // 429 Too Many Requests
    if resp.StatusCode == 429 {
        return true
    }
    return false
}

// waitForRetry 等待重试
func (e *Executor) waitForRetry(ctx context.Context, attempt int) {
    // 指数退避
    delay := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
    if delay > 10*time.Second {
        delay = 10 * time.Second
    }

    select {
    case <-ctx.Done():
    case <-time.After(delay):
    }
}

// isSuccessStatus 判断是否为成功状态
func (e *Executor) isSuccessStatus(statusCode int, expect200OKWithError bool) bool {
    if expect200OKWithError {
        return false // 需要检查响应体
    }
    return statusCode >= 200 && statusCode < 300
}

// LocationCache 位置缓存接口
type LocationCache interface {
    Get(bucketName string) (string, bool)
    Set(bucketName, location string)
    Delete(bucketName string)
}

// closeResponse 关闭响应
func closeResponse(resp *http.Response) {
    if resp != nil && resp.Body != nil {
        io.Copy(io.Discard, resp.Body)
        resp.Body.Close()
    }
}
```

#### 1.4.3 创建 `internal/core/response.go`

```go
// internal/core/response.go
package core

import (
    "encoding/xml"
    "io"
    "net/http"
    "strconv"
    "time"

    "github.com/Scorpio69t/rustfs-go/errors"
    "github.com/Scorpio69t/rustfs-go/types"
)

// ResponseParser 响应解析器
type ResponseParser struct{}

// NewResponseParser 创建响应解析器
func NewResponseParser() *ResponseParser {
    return &ResponseParser{}
}

// ParseXML 解析 XML 响应
func (p *ResponseParser) ParseXML(resp *http.Response, v interface{}) error {
    if resp.Body == nil {
        return errors.NewAPIError(errors.ErrCodeInternalError, "empty response body", resp.StatusCode)
    }
    defer resp.Body.Close()

    return xml.NewDecoder(resp.Body).Decode(v)
}

// ParseObjectInfo 从响应头解析对象信息
func (p *ResponseParser) ParseObjectInfo(resp *http.Response, bucketName, objectName string) (types.ObjectInfo, error) {
    header := resp.Header

    info := types.ObjectInfo{
        Key:         objectName,
        ContentType: header.Get("Content-Type"),
        ETag:        trimETag(header.Get("ETag")),
    }

    // 解析 Content-Length
    if cl := header.Get("Content-Length"); cl != "" {
        if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
            info.Size = size
        }
    }

    // 解析 Last-Modified
    if lm := header.Get("Last-Modified"); lm != "" {
        if t, err := time.Parse(http.TimeFormat, lm); err == nil {
            info.LastModified = t
        }
    }

    // 解析版本信息
    info.VersionID = header.Get("x-amz-version-id")
    info.IsDeleteMarker = header.Get("x-amz-delete-marker") == "true"

    // 解析存储类
    info.StorageClass = header.Get("x-amz-storage-class")

    // 解析复制状态
    info.ReplicationStatus = header.Get("x-amz-replication-status")

    // 解析用户元数据
    info.UserMetadata = make(types.StringMap)
    for k, v := range header {
        if len(k) > len("X-Amz-Meta-") && k[:len("X-Amz-Meta-")] == "X-Amz-Meta-" {
            info.UserMetadata[k[len("X-Amz-Meta-"):]] = v[0]
        }
    }

    // 解析标签数量
    if tc := header.Get("x-amz-tagging-count"); tc != "" {
        if count, err := strconv.Atoi(tc); err == nil {
            info.UserTagCount = count
        }
    }

    // 解析校验和
    info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
    info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
    info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
    info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
    info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

    return info, nil
}

// ParseUploadInfo 从响应解析上传信息
func (p *ResponseParser) ParseUploadInfo(resp *http.Response, bucketName, objectName string) (types.UploadInfo, error) {
    header := resp.Header

    info := types.UploadInfo{
        Bucket:    bucketName,
        Key:       objectName,
        ETag:      trimETag(header.Get("ETag")),
        VersionID: header.Get("x-amz-version-id"),
    }

    // 解析校验和
    info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
    info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
    info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
    info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
    info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

    return info, nil
}

// ParseError 解析错误响应
func (p *ResponseParser) ParseError(resp *http.Response, bucketName, objectName string) error {
    return errors.ParseErrorResponse(resp, bucketName, objectName)
}

// trimETag 去除 ETag 的引号
func trimETag(etag string) string {
    if len(etag) > 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
        return etag[1 : len(etag)-1]
    }
    return etag
}
```

#### 验证清单
- [ ] `internal/core/request.go` 已创建
- [ ] `internal/core/executor.go` 已创建
- [ ] `internal/core/response.go` 已创建
- [ ] 编译通过
- [ ] 与现有代码集成测试

---

### 任务 1.5：创建内部缓存包 `internal/cache/`
**状态**: ⬜ 未开始  
**预计时间**: 0.5 天

#### 1.5.1 创建 `internal/cache/location.go`

```go
// internal/cache/location.go
package cache

import (
    "sync"
    "time"
)

// LocationCache 桶位置缓存
type LocationCache struct {
    mu      sync.RWMutex
    entries map[string]locationEntry
    ttl     time.Duration
}

type locationEntry struct {
    location  string
    expiresAt time.Time
}

// NewLocationCache 创建位置缓存
func NewLocationCache(ttl time.Duration) *LocationCache {
    if ttl <= 0 {
        ttl = 5 * time.Minute
    }
    return &LocationCache{
        entries: make(map[string]locationEntry),
        ttl:     ttl,
    }
}

// Get 获取桶位置
func (c *LocationCache) Get(bucketName string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.entries[bucketName]
    if !ok {
        return "", false
    }

    if time.Now().After(entry.expiresAt) {
        return "", false
    }

    return entry.location, true
}

// Set 设置桶位置
func (c *LocationCache) Set(bucketName, location string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries[bucketName] = locationEntry{
        location:  location,
        expiresAt: time.Now().Add(c.ttl),
    }
}

// Delete 删除桶位置
func (c *LocationCache) Delete(bucketName string) {
    c.mu.Lock()
    defer c.mu.Unlock()

    delete(c.entries, bucketName)
}

// Clear 清空缓存
func (c *LocationCache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries = make(map[string]locationEntry)
}
```

---

### 任务 1.6：更新根目录客户端文件
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 1.6.1 创建新的 `options.go`

```go
// options.go
package rustfs

import (
    "net/http"
    "net/http/httptrace"
    "net/url"

    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
    "github.com/Scorpio69t/rustfs-go/types"
)

// Options 客户端配置选项
type Options struct {
    // Credentials 凭证提供者
    // 必需，用于签名请求
    Credentials *credentials.Credentials

    // Secure 是否使用 HTTPS
    // 默认: false
    Secure bool

    // Region 区域
    // 如果不设置，将自动检测
    Region string

    // Transport 自定义 HTTP 传输
    // 如果不设置，使用默认传输
    Transport http.RoundTripper

    // Trace HTTP 追踪客户端
    Trace *httptrace.ClientTrace

    // BucketLookup 桶查找类型
    // 默认: BucketLookupAuto
    BucketLookup types.BucketLookupType

    // CustomRegionViaURL 自定义区域查找函数
    CustomRegionViaURL func(u url.URL) string

    // BucketLookupViaURL 自定义桶查找函数
    BucketLookupViaURL func(u url.URL, bucketName string) types.BucketLookupType

    // TrailingHeaders 启用尾部头（用于流式上传）
    // 需要服务器支持
    TrailingHeaders bool

    // MaxRetries 最大重试次数
    // 默认: 10，设置为 1 禁用重试
    MaxRetries int
}

// validate 验证选项
func (o *Options) validate() error {
    if o == nil {
        return errInvalidArgument("options cannot be nil")
    }
    if o.Credentials == nil {
        return errInvalidArgument("credentials are required")
    }
    return nil
}

// setDefaults 设置默认值
func (o *Options) setDefaults() {
    if o.MaxRetries <= 0 {
        o.MaxRetries = 10
    }
    if o.BucketLookup == 0 {
        o.BucketLookup = types.BucketLookupAuto
    }
}

// errInvalidArgument 创建无效参数错误
func errInvalidArgument(message string) error {
    return &invalidArgumentError{message: message}
}

type invalidArgumentError struct {
    message string
}

func (e *invalidArgumentError) Error() string {
    return e.message
}
```

---

## 🔧 第二阶段：核心模块实现（预计 7 天）

### 任务 2.1：实现签名模块 `internal/signer/`
**状态**: ⬜ 未开始  
**预计时间**: 2 天

> **注意**: 此任务主要是将现有的 `pkg/signer/` 逻辑迁移到内部包，并进行适当的封装。

#### 2.1.1 创建 `internal/signer/signer.go`

```go
// internal/signer/signer.go
package signer

import (
    "net/http"
    "time"

    "github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

// Signer 签名器接口
type Signer interface {
    // Sign 签名请求
    Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request

    // Presign 预签名请求
    Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request
}

// SignerType 签名类型
type SignerType int

const (
    SignerV4 SignerType = iota
    SignerV2
    SignerAnonymous
)

// NewSigner 创建签名器
func NewSigner(signerType SignerType) Signer {
    switch signerType {
    case SignerV2:
        return &V2Signer{}
    case SignerAnonymous:
        return &AnonymousSigner{}
    default:
        return &V4Signer{}
    }
}

// SignRequest 签名请求的便捷函数
func SignRequest(req *http.Request, creds credentials.Value, region string) *http.Request {
    signer := NewSigner(getSignerType(creds.SignerType))
    return signer.Sign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region)
}

func getSignerType(st credentials.SignatureType) SignerType {
    switch st {
    case credentials.SignatureV2:
        return SignerV2
    case credentials.SignatureAnonymous:
        return SignerAnonymous
    default:
        return SignerV4
    }
}
```

#### 2.1.2 创建 `internal/signer/v4.go`（V4 签名实现）

```go
// internal/signer/v4.go
package signer

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "net/http"
    "sort"
    "strings"
    "time"
)

// V4Signer AWS Signature Version 4 签名器
type V4Signer struct{}

// Sign 使用 V4 算法签名请求
func (s *V4Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
    // 设置时间
    t := time.Now().UTC()
    req.Header.Set("X-Amz-Date", t.Format("20060102T150405Z"))

    // 设置 session token
    if sessionToken != "" {
        req.Header.Set("X-Amz-Security-Token", sessionToken)
    }

    // 计算签名
    signature := s.calculateSignature(req, accessKey, secretKey, region, t)

    // 构建 Authorization 头
    auth := s.buildAuthorizationHeader(req, accessKey, region, signature, t)
    req.Header.Set("Authorization", auth)

    return req
}

// Presign 使用 V4 算法预签名请求
func (s *V4Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
    // TODO: 实现预签名逻辑
    return req
}

// calculateSignature 计算签名
func (s *V4Signer) calculateSignature(req *http.Request, accessKey, secretKey, region string, t time.Time) string {
    // 1. 创建规范请求
    canonicalRequest := s.createCanonicalRequest(req)

    // 2. 创建待签名字符串
    stringToSign := s.createStringToSign(canonicalRequest, region, t)

    // 3. 计算签名
    signingKey := s.deriveSigningKey(secretKey, region, t)
    signature := hmacSHA256(signingKey, []byte(stringToSign))

    return hex.EncodeToString(signature)
}

// createCanonicalRequest 创建规范请求
func (s *V4Signer) createCanonicalRequest(req *http.Request) string {
    // HTTP Method
    method := req.Method

    // Canonical URI
    uri := req.URL.Path
    if uri == "" {
        uri = "/"
    }

    // Canonical Query String
    queryString := req.URL.Query().Encode()

    // Canonical Headers
    headers, signedHeaders := s.canonicalHeaders(req.Header)

    // Payload Hash
    payloadHash := req.Header.Get("X-Amz-Content-Sha256")
    if payloadHash == "" {
        payloadHash = "UNSIGNED-PAYLOAD"
    }

    return strings.Join([]string{
        method,
        uri,
        queryString,
        headers,
        signedHeaders,
        payloadHash,
    }, "\n")
}

// canonicalHeaders 创建规范头部
func (s *V4Signer) canonicalHeaders(header http.Header) (canonical, signed string) {
    var keys []string
    for k := range header {
        keys = append(keys, strings.ToLower(k))
    }
    sort.Strings(keys)

    var headers []string
    var signedHeaders []string
    for _, k := range keys {
        if k == "host" || strings.HasPrefix(k, "x-amz-") || k == "content-type" {
            headers = append(headers, k+":"+strings.TrimSpace(header.Get(k)))
            signedHeaders = append(signedHeaders, k)
        }
    }

    return strings.Join(headers, "\n") + "\n", strings.Join(signedHeaders, ";")
}

// createStringToSign 创建待签名字符串
func (s *V4Signer) createStringToSign(canonicalRequest, region string, t time.Time) string {
    scope := s.credentialScope(region, t)
    hash := sha256.Sum256([]byte(canonicalRequest))
    return strings.Join([]string{
        "AWS4-HMAC-SHA256",
        t.Format("20060102T150405Z"),
        scope,
        hex.EncodeToString(hash[:]),
    }, "\n")
}

// credentialScope 创建凭证范围
func (s *V4Signer) credentialScope(region string, t time.Time) string {
    return strings.Join([]string{
        t.Format("20060102"),
        region,
        "s3",
        "aws4_request",
    }, "/")
}

// deriveSigningKey 派生签名密钥
func (s *V4Signer) deriveSigningKey(secretKey, region string, t time.Time) []byte {
    dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(t.Format("20060102")))
    regionKey := hmacSHA256(dateKey, []byte(region))
    serviceKey := hmacSHA256(regionKey, []byte("s3"))
    signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
    return signingKey
}

// buildAuthorizationHeader 构建 Authorization 头
func (s *V4Signer) buildAuthorizationHeader(req *http.Request, accessKey, region, signature string, t time.Time) string {
    _, signedHeaders := s.canonicalHeaders(req.Header)
    scope := s.credentialScope(region, t)
    
    return "AWS4-HMAC-SHA256 " +
        "Credential=" + accessKey + "/" + scope + ", " +
        "SignedHeaders=" + signedHeaders + ", " +
        "Signature=" + signature
}

// hmacSHA256 计算 HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
    h := hmac.New(sha256.New, key)
    h.Write(data)
    return h.Sum(nil)
}
```

---

### 任务 2.2：实现传输层 `internal/transport/`
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 2.2.1 创建 `internal/transport/transport.go`

```go
// internal/transport/transport.go
package transport

import (
    "crypto/tls"
    "net"
    "net/http"
    "time"
)

// DefaultTransport 创建默认 HTTP 传输
func DefaultTransport(secure bool) (*http.Transport, error) {
    tr := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        MaxIdleConns:          100,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        // 禁用压缩以支持 Range 请求
        DisableCompression: true,
    }

    if secure {
        tr.TLSClientConfig = &tls.Config{
            MinVersion: tls.VersionTLS12,
        }
    }

    return tr, nil
}

// TransportOptions 传输选项
type TransportOptions struct {
    // TLS 配置
    TLSConfig *tls.Config
    
    // 超时设置
    DialTimeout   time.Duration
    DialKeepAlive time.Duration
    
    // 连接池
    MaxIdleConns        int
    MaxIdleConnsPerHost int
    IdleConnTimeout     time.Duration
    
    // 代理
    Proxy func(*http.Request) (*url.URL, error)
}

// NewTransport 创建自定义传输
func NewTransport(opts TransportOptions) *http.Transport {
    dialTimeout := opts.DialTimeout
    if dialTimeout <= 0 {
        dialTimeout = 30 * time.Second
    }
    
    dialKeepAlive := opts.DialKeepAlive
    if dialKeepAlive <= 0 {
        dialKeepAlive = 30 * time.Second
    }
    
    maxIdleConns := opts.MaxIdleConns
    if maxIdleConns <= 0 {
        maxIdleConns = 100
    }
    
    idleConnTimeout := opts.IdleConnTimeout
    if idleConnTimeout <= 0 {
        idleConnTimeout = 90 * time.Second
    }

    tr := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   dialTimeout,
            KeepAlive: dialKeepAlive,
        }).DialContext,
        MaxIdleConns:          maxIdleConns,
        MaxIdleConnsPerHost:   opts.MaxIdleConnsPerHost,
        IdleConnTimeout:       idleConnTimeout,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        DisableCompression:    true,
    }
    
    if opts.TLSConfig != nil {
        tr.TLSClientConfig = opts.TLSConfig
    }
    
    if opts.Proxy != nil {
        tr.Proxy = opts.Proxy
    } else {
        tr.Proxy = http.ProxyFromEnvironment
    }

    return tr
}
```

---

### 任务 2.3：创建服务接口定义
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 2.3.1 创建 `bucket/service.go`

```go
// bucket/service.go
package bucket

import (
    "context"

    "github.com/Scorpio69t/rustfs-go/pkg/cors"
    "github.com/Scorpio69t/rustfs-go/pkg/lifecycle"
    "github.com/Scorpio69t/rustfs-go/pkg/notification"
    "github.com/Scorpio69t/rustfs-go/pkg/policy"
    "github.com/Scorpio69t/rustfs-go/pkg/replication"
    "github.com/Scorpio69t/rustfs-go/pkg/tags"
    "github.com/Scorpio69t/rustfs-go/types"
)

// Service 桶服务接口
type Service interface {
    // 基础操作
    Create(ctx context.Context, name string, opts ...CreateOption) error
    Delete(ctx context.Context, name string, opts ...DeleteOption) error
    Exists(ctx context.Context, name string) (bool, error)
    List(ctx context.Context) ([]types.BucketInfo, error)

    // 配置管理
    Config() ConfigService

    // 策略管理
    Policy() PolicyService
}

// ConfigService 桶配置服务接口
type ConfigService interface {
    // 生命周期
    SetLifecycle(ctx context.Context, bucket string, config *lifecycle.Configuration) error
    GetLifecycle(ctx context.Context, bucket string) (*lifecycle.Configuration, error)
    DeleteLifecycle(ctx context.Context, bucket string) error

    // 版本控制
    SetVersioning(ctx context.Context, bucket string, config types.VersioningConfig) error
    GetVersioning(ctx context.Context, bucket string) (types.VersioningConfig, error)

    // CORS
    SetCORS(ctx context.Context, bucket string, config *cors.Config) error
    GetCORS(ctx context.Context, bucket string) (*cors.Config, error)
    DeleteCORS(ctx context.Context, bucket string) error

    // 标签
    SetTags(ctx context.Context, bucket string, t *tags.Tags) error
    GetTags(ctx context.Context, bucket string) (*tags.Tags, error)
    DeleteTags(ctx context.Context, bucket string) error

    // 加密
    SetEncryption(ctx context.Context, bucket string, config *EncryptionConfig) error
    GetEncryption(ctx context.Context, bucket string) (*EncryptionConfig, error)
    DeleteEncryption(ctx context.Context, bucket string) error

    // 复制
    SetReplication(ctx context.Context, bucket string, config *replication.Config) error
    GetReplication(ctx context.Context, bucket string) (*replication.Config, error)
    DeleteReplication(ctx context.Context, bucket string) error

    // 通知
    SetNotification(ctx context.Context, bucket string, config notification.Configuration) error
    GetNotification(ctx context.Context, bucket string) (notification.Configuration, error)
}

// PolicyService 桶策略服务接口
type PolicyService interface {
    Set(ctx context.Context, bucket string, policy *policy.BucketPolicy) error
    Get(ctx context.Context, bucket string) (*policy.BucketPolicy, error)
    Delete(ctx context.Context, bucket string) error
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
    // SSE 算法: AES256 或 aws:kms
    Algorithm string
    // KMS Key ID (仅当 Algorithm 为 aws:kms 时使用)
    KMSKeyID string
}

// CreateOption 创建桶选项
type CreateOption func(*CreateOptions)

// CreateOptions 创建桶选项结构
type CreateOptions struct {
    Region       string
    ObjectLock   bool
    Tags         map[string]string
}

// WithRegion 设置区域
func WithRegion(region string) CreateOption {
    return func(o *CreateOptions) {
        o.Region = region
    }
}

// WithObjectLock 启用对象锁定
func WithObjectLock(enabled bool) CreateOption {
    return func(o *CreateOptions) {
        o.ObjectLock = enabled
    }
}

// WithTags 设置标签
func WithTags(tags map[string]string) CreateOption {
    return func(o *CreateOptions) {
        o.Tags = tags
    }
}

// DeleteOption 删除桶选项
type DeleteOption func(*DeleteOptions)

// DeleteOptions 删除桶选项结构
type DeleteOptions struct {
    ForceDelete bool
}

// WithForceDelete 强制删除（包括所有对象）
func WithForceDelete(force bool) DeleteOption {
    return func(o *DeleteOptions) {
        o.ForceDelete = force
    }
}
```

#### 2.3.2 创建 `object/service.go`

```go
// object/service.go
package object

import (
    "context"
    "io"
    "iter"
    "time"

    "github.com/Scorpio69t/rustfs-go/pkg/encrypt"
    "github.com/Scorpio69t/rustfs-go/pkg/tags"
    "github.com/Scorpio69t/rustfs-go/types"
)

// Service 对象服务接口
type Service interface {
    // 上传服务
    Upload() UploadService

    // 下载服务
    Download() DownloadService

    // 基础操作
    Stat(ctx context.Context, bucket, key string, opts ...StatOption) (types.ObjectInfo, error)
    Delete(ctx context.Context, bucket, key string, opts ...DeleteOption) error
    DeleteMultiple(ctx context.Context, bucket string, objects []types.ObjectToDelete, opts ...DeleteOption) ([]types.DeletedObject, []types.DeleteError, error)
    Copy(ctx context.Context, dest CopyDestination, src CopySource, opts ...CopyOption) (types.UploadInfo, error)

    // 标签操作
    SetTags(ctx context.Context, bucket, key string, t *tags.Tags, opts ...TagOption) error
    GetTags(ctx context.Context, bucket, key string, opts ...TagOption) (*tags.Tags, error)
    DeleteTags(ctx context.Context, bucket, key string, opts ...TagOption) error

    // 列表操作
    List(ctx context.Context, bucket string, opts ...ListOption) <-chan types.ObjectInfo
    ListIter(ctx context.Context, bucket string, opts ...ListOption) iter.Seq[types.ObjectInfo]
}

// UploadService 上传服务接口
type UploadService interface {
    // Put 上传对象
    Put(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...PutOption) (types.UploadInfo, error)

    // PutFile 从文件上传
    PutFile(ctx context.Context, bucket, key, filePath string, opts ...PutOption) (types.UploadInfo, error)

    // Multipart 分片上传服务
    Multipart() MultipartService
}

// DownloadService 下载服务接口
type DownloadService interface {
    // Get 获取对象
    Get(ctx context.Context, bucket, key string, opts ...GetOption) (*Object, error)

    // GetFile 下载到文件
    GetFile(ctx context.Context, bucket, key, filePath string, opts ...GetOption) error

    // GetRange 范围下载
    GetRange(ctx context.Context, bucket, key string, offset, length int64, opts ...GetOption) (*Object, error)
}

// MultipartService 分片上传服务接口
type MultipartService interface {
    // Create 创建分片上传
    Create(ctx context.Context, bucket, key string, opts ...PutOption) (string, error)

    // UploadPart 上传分片
    UploadPart(ctx context.Context, bucket, key, uploadID string, partNumber int, reader io.Reader, size int64, opts ...PartOption) (types.PartInfo, error)

    // Complete 完成分片上传
    Complete(ctx context.Context, bucket, key, uploadID string, parts []types.CompletePart, opts ...PutOption) (types.UploadInfo, error)

    // Abort 中止分片上传
    Abort(ctx context.Context, bucket, key, uploadID string) error

    // ListParts 列出已上传分片
    ListParts(ctx context.Context, bucket, key, uploadID string, opts ...ListPartOption) ([]types.PartInfo, error)

    // ListUploads 列出进行中的分片上传
    ListUploads(ctx context.Context, bucket string, opts ...ListUploadOption) <-chan types.MultipartInfo
}

// Object 下载对象封装
type Object struct {
    io.ReadCloser
    info types.ObjectInfo
}

// Info 返回对象信息
func (o *Object) Info() types.ObjectInfo {
    return o.info
}

// CopySource 复制源
type CopySource struct {
    Bucket    string
    Key       string
    VersionID string
}

// CopyDestination 复制目标
type CopyDestination struct {
    Bucket string
    Key    string
}

// PutOption 上传选项
type PutOption func(*PutOptions)

// PutOptions 上传选项结构
type PutOptions struct {
    ContentType        string
    ContentEncoding    string
    ContentDisposition string
    ContentLanguage    string
    CacheControl       string
    Expires            time.Time
    Metadata           map[string]string
    Tags               map[string]string
    StorageClass       string
    SSE                encrypt.ServerSide
    RetentionMode      types.RetentionMode
    RetainUntilDate    time.Time
    LegalHold          types.LegalHoldStatus
    PartSize           uint64
    NumThreads         uint
    DisableMultipart   bool
    Checksum           types.ChecksumType
    SendContentMD5     bool
    Progress           func(uploaded, total int64)
}

// GetOption 下载选项
type GetOption func(*GetOptions)

// GetOptions 下载选项结构
type GetOptions struct {
    VersionID         string
    SSE               encrypt.ServerSide
    IfMatch           string
    IfNoneMatch       string
    IfModifiedSince   time.Time
    IfUnmodifiedSince time.Time
    RangeStart        int64
    RangeEnd          int64
}

// StatOption 状态查询选项
type StatOption func(*StatOptions)

// StatOptions 状态查询选项结构
type StatOptions struct {
    VersionID string
    SSE       encrypt.ServerSide
}

// DeleteOption 删除选项
type DeleteOption func(*DeleteOptions)

// DeleteOptions 删除选项结构
type DeleteOptions struct {
    VersionID        string
    GovernanceBypass bool
}

// CopyOption 复制选项
type CopyOption func(*CopyOptions)

// CopyOptions 复制选项结构
type CopyOptions struct {
    // 源对象条件
    IfMatch           string
    IfNoneMatch       string
    IfModifiedSince   time.Time
    IfUnmodifiedSince time.Time

    // 目标对象设置
    ContentType        string
    ContentEncoding    string
    ContentDisposition string
    Metadata           map[string]string
    Tags               map[string]string
    StorageClass       string
    SSE                encrypt.ServerSide

    // 元数据处理
    MetadataDirective string // COPY 或 REPLACE
    TaggingDirective  string // COPY 或 REPLACE
}

// ListOption 列表选项
type ListOption func(*ListOptions)

// ListOptions 列表选项结构
type ListOptions struct {
    Prefix       string
    Delimiter    string
    StartAfter   string
    MaxKeys      int
    Recursive    bool
    WithVersions bool
    WithMetadata bool
}

// TagOption 标签操作选项
type TagOption func(*TagOptions)

// TagOptions 标签选项结构
type TagOptions struct {
    VersionID string
}

// PartOption 分片上传选项
type PartOption func(*PartOptions)

// PartOptions 分片上传选项结构
type PartOptions struct {
    SSE        encrypt.ServerSide
    ContentMD5 string
    Checksum   types.ChecksumType
}

// ListPartOption 列出分片选项
type ListPartOption func(*ListPartOptions)

// ListPartOptions 列出分片选项结构
type ListPartOptions struct {
    PartNumberMarker int
    MaxParts         int
}

// ListUploadOption 列出上传选项
type ListUploadOption func(*ListUploadOptions)

// ListUploadOptions 列出上传选项结构
type ListUploadOptions struct {
    Prefix         string
    KeyMarker      string
    UploadIDMarker string
    Delimiter      string
    MaxUploads     int
}
```

---

### 任务 2.4：实现选项函数
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 2.4.1 创建 `object/options.go`

```go
// object/options.go
package object

import (
    "time"

    "github.com/Scorpio69t/rustfs-go/pkg/encrypt"
    "github.com/Scorpio69t/rustfs-go/types"
)

// ========== Put 选项 ==========

// WithContentType 设置 Content-Type
func WithContentType(ct string) PutOption {
    return func(o *PutOptions) {
        o.ContentType = ct
    }
}

// WithContentEncoding 设置 Content-Encoding
func WithContentEncoding(ce string) PutOption {
    return func(o *PutOptions) {
        o.ContentEncoding = ce
    }
}

// WithContentDisposition 设置 Content-Disposition
func WithContentDisposition(cd string) PutOption {
    return func(o *PutOptions) {
        o.ContentDisposition = cd
    }
}

// WithContentLanguage 设置 Content-Language
func WithContentLanguage(cl string) PutOption {
    return func(o *PutOptions) {
        o.ContentLanguage = cl
    }
}

// WithCacheControl 设置 Cache-Control
func WithCacheControl(cc string) PutOption {
    return func(o *PutOptions) {
        o.CacheControl = cc
    }
}

// WithExpires 设置过期时间
func WithExpires(exp time.Time) PutOption {
    return func(o *PutOptions) {
        o.Expires = exp
    }
}

// WithMetadata 设置用户元数据
func WithMetadata(meta map[string]string) PutOption {
    return func(o *PutOptions) {
        o.Metadata = meta
    }
}

// WithTags 设置标签
func WithTags(tags map[string]string) PutOption {
    return func(o *PutOptions) {
        o.Tags = tags
    }
}

// WithStorageClass 设置存储类
func WithStorageClass(sc string) PutOption {
    return func(o *PutOptions) {
        o.StorageClass = sc
    }
}

// WithServerSideEncryption 设置服务端加密
func WithServerSideEncryption(sse encrypt.ServerSide) PutOption {
    return func(o *PutOptions) {
        o.SSE = sse
    }
}

// WithRetention 设置对象保留
func WithRetention(mode types.RetentionMode, until time.Time) PutOption {
    return func(o *PutOptions) {
        o.RetentionMode = mode
        o.RetainUntilDate = until
    }
}

// WithLegalHold 设置法律保留
func WithLegalHold(status types.LegalHoldStatus) PutOption {
    return func(o *PutOptions) {
        o.LegalHold = status
    }
}

// WithPartSize 设置分片大小
func WithPartSize(size uint64) PutOption {
    return func(o *PutOptions) {
        o.PartSize = size
    }
}

// WithNumThreads 设置并发线程数
func WithNumThreads(n uint) PutOption {
    return func(o *PutOptions) {
        o.NumThreads = n
    }
}

// WithDisableMultipart 禁用分片上传
func WithDisableMultipart(disable bool) PutOption {
    return func(o *PutOptions) {
        o.DisableMultipart = disable
    }
}

// WithChecksum 设置校验和类型
func WithChecksum(ct types.ChecksumType) PutOption {
    return func(o *PutOptions) {
        o.Checksum = ct
    }
}

// WithProgress 设置进度回调
func WithProgress(fn func(uploaded, total int64)) PutOption {
    return func(o *PutOptions) {
        o.Progress = fn
    }
}

// ========== Get 选项 ==========

// WithVersionID 设置版本 ID
func WithVersionID(vid string) GetOption {
    return func(o *GetOptions) {
        o.VersionID = vid
    }
}

// WithSSE 设置服务端加密（用于解密）
func WithSSE(sse encrypt.ServerSide) GetOption {
    return func(o *GetOptions) {
        o.SSE = sse
    }
}

// WithIfMatch 设置 If-Match 条件
func WithIfMatch(etag string) GetOption {
    return func(o *GetOptions) {
        o.IfMatch = etag
    }
}

// WithIfNoneMatch 设置 If-None-Match 条件
func WithIfNoneMatch(etag string) GetOption {
    return func(o *GetOptions) {
        o.IfNoneMatch = etag
    }
}

// WithIfModifiedSince 设置 If-Modified-Since 条件
func WithIfModifiedSince(t time.Time) GetOption {
    return func(o *GetOptions) {
        o.IfModifiedSince = t
    }
}

// WithIfUnmodifiedSince 设置 If-Unmodified-Since 条件
func WithIfUnmodifiedSince(t time.Time) GetOption {
    return func(o *GetOptions) {
        o.IfUnmodifiedSince = t
    }
}

// WithRange 设置范围下载
func WithRange(start, end int64) GetOption {
    return func(o *GetOptions) {
        o.RangeStart = start
        o.RangeEnd = end
    }
}

// ========== List 选项 ==========

// WithPrefix 设置前缀过滤
func WithPrefix(prefix string) ListOption {
    return func(o *ListOptions) {
        o.Prefix = prefix
    }
}

// WithDelimiter 设置分隔符
func WithDelimiter(delimiter string) ListOption {
    return func(o *ListOptions) {
        o.Delimiter = delimiter
    }
}

// WithStartAfter 设置起始键
func WithStartAfter(key string) ListOption {
    return func(o *ListOptions) {
        o.StartAfter = key
    }
}

// WithMaxKeys 设置最大返回数量
func WithMaxKeys(max int) ListOption {
    return func(o *ListOptions) {
        o.MaxKeys = max
    }
}

// WithRecursive 递归列出
func WithRecursive(recursive bool) ListOption {
    return func(o *ListOptions) {
        o.Recursive = recursive
    }
}

// WithVersions 包含版本
func WithVersions(include bool) ListOption {
    return func(o *ListOptions) {
        o.WithVersions = include
    }
}

// WithObjectMetadata 包含元数据
func WithObjectMetadata(include bool) ListOption {
    return func(o *ListOptions) {
        o.WithMetadata = include
    }
}

// ========== Delete 选项 ==========

// WithDeleteVersionID 设置删除版本
func WithDeleteVersionID(vid string) DeleteOption {
    return func(o *DeleteOptions) {
        o.VersionID = vid
    }
}

// WithGovernanceBypass 绕过治理模式
func WithGovernanceBypass(bypass bool) DeleteOption {
    return func(o *DeleteOptions) {
        o.GovernanceBypass = bypass
    }
}

// ========== Stat 选项 ==========

// WithStatVersionID 设置查询版本
func WithStatVersionID(vid string) StatOption {
    return func(o *StatOptions) {
        o.VersionID = vid
    }
}

// WithStatSSE 设置加密（用于解密元数据）
func WithStatSSE(sse encrypt.ServerSide) StatOption {
    return func(o *StatOptions) {
        o.SSE = sse
    }
}
```

---

## 📦 第三阶段：Bucket 模块实现（预计 5 天）

### 任务 3.1：实现 Bucket 基础操作
**状态**: ⬜ 未开始  
**预计时间**: 2 天

> 将 `api-put-bucket.go`、`api-remove.go`、`api-stat.go`、`api-list.go` 中的桶操作迁移到 `bucket/` 包

#### 待迁移的功能
- [ ] `MakeBucket` → `bucket.Create`
- [ ] `RemoveBucket` → `bucket.Delete`
- [ ] `BucketExists` → `bucket.Exists`
- [ ] `ListBuckets` → `bucket.List`
- [ ] `GetBucketLocation` → 内部使用

### 任务 3.2：实现 Bucket 配置操作
**状态**: ⬜ 未开始  
**预计时间**: 2 天

> 将各 `api-bucket-*.go` 文件迁移到 `bucket/config/` 包

#### 待迁移的功能
- [ ] `api-bucket-lifecycle.go` → `bucket/config/lifecycle.go`
- [ ] `api-bucket-versioning.go` → `bucket/config/versioning.go`
- [ ] `api-bucket-cors.go` → `bucket/config/cors.go`
- [ ] `api-bucket-encryption.go` → `bucket/config/encryption.go`
- [ ] `api-bucket-tagging.go` → `bucket/config/tagging.go`
- [ ] `api-bucket-replication.go` → `bucket/config/replication.go`
- [ ] `api-bucket-notification.go` → `bucket/config/notification.go`
- [ ] `api-bucket-qos.go` → `bucket/config/qos.go`

### 任务 3.3：实现 Bucket 策略操作
**状态**: ⬜ 未开始  
**预计时间**: 1 天

> 将 `api-bucket-policy.go` 迁移到 `bucket/policy/` 包

---

## 📁 第四阶段：Object 模块实现（预计 8 天）

### 任务 4.1：实现上传功能 `object/upload/`
**状态**: ⬜ 未开始  
**预计时间**: 3 天

#### 待迁移的功能
- [ ] `api-put-object.go` → `object/upload/simple.go`
- [ ] `api-put-object-multipart.go` → `object/upload/multipart.go`
- [ ] `api-put-object-streaming.go` → `object/upload/streaming.go`
- [ ] `api-put-object-file-context.go` → `object/upload/file.go`
- [ ] `api-put-object-common.go` → `object/upload/common.go`
- [ ] `api-append-object.go` → `object/upload/append.go`

### 任务 4.2：实现下载功能 `object/download/`
**状态**: ⬜ 未开始  
**预计时间**: 2 天

#### 待迁移的功能
- [ ] `api-get-object.go` → `object/download/simple.go`
- [ ] `api-get-object-file.go` → `object/download/file.go`
- [ ] Range 下载 → `object/download/range.go`

### 任务 4.3：实现对象管理功能 `object/manage/`
**状态**: ⬜ 未开始  
**预计时间**: 2 天

#### 待迁移的功能
- [ ] `api-stat.go` (StatObject) → `object/manage/stat.go`
- [ ] `api-remove.go` (RemoveObject) → `object/manage/delete.go`
- [ ] `api-copy-object.go` → `object/manage/copy.go`
- [ ] `api-compose-object.go` → `object/manage/compose.go`
- [ ] `api-object-tagging.go` → `object/manage/tagging.go`
- [ ] `api-list.go` (ListObjects) → `object/manage/list.go`
- [ ] `api-restore.go` → `object/manage/restore.go`

### 任务 4.4：实现预签名功能 `object/presign/`
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 待迁移的功能
- [ ] `api-presigned.go` → `object/presign/`

---

## 🔗 第五阶段：兼容层和测试（预计 5 天）

### 任务 5.1：创建兼容层
**状态**: ⬜ 未开始  
**预计时间**: 2 天

在根包中创建兼容旧 API 的方法，标记为 deprecated。

```go
// compat.go
package rustfs

import (
    "context"
    "io"

    "github.com/Scorpio69t/rustfs-go/object"
    "github.com/Scorpio69t/rustfs-go/types"
)

// Deprecated: Use Client.Object().Upload().Put instead.
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts PutObjectOptions) (UploadInfo, error) {
    // 转换选项
    var putOpts []object.PutOption
    if opts.ContentType != "" {
        putOpts = append(putOpts, object.WithContentType(opts.ContentType))
    }
    if len(opts.UserMetadata) > 0 {
        putOpts = append(putOpts, object.WithMetadata(opts.UserMetadata))
    }
    // ... 更多选项转换

    info, err := c.Object().Upload().Put(ctx, bucketName, objectName, reader, objectSize, putOpts...)
    if err != nil {
        return UploadInfo{}, err
    }

    // 转换返回类型
    return UploadInfo{
        Bucket:   info.Bucket,
        Key:      info.Key,
        ETag:     info.ETag,
        Size:     info.Size,
        // ...
    }, nil
}

// Deprecated: Use Client.Object().Download().Get instead.
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (*Object, error) {
    // ... 兼容实现
    return nil, nil
}
```

### 任务 5.2：编写单元测试
**状态**: ⬜ 未开始  
**预计时间**: 2 天

#### 测试清单
- [ ] `types/` 包测试
- [ ] `errors/` 包测试
- [ ] `internal/core/` 包测试
- [ ] `internal/cache/` 包测试
- [ ] `bucket/` 包测试
- [ ] `object/` 包测试

### 任务 5.3：更新文档和示例
**状态**: ⬜ 未开始  
**预计时间**: 1 天

#### 文档清单
- [ ] 更新 README.md
- [ ] 创建 docs/getting-started.md
- [ ] 创建 docs/migration-guide.md
- [ ] 创建 examples/basic/ 示例
- [ ] 创建 examples/advanced/ 示例

---

## 📝 附录

### A. 检查清单模板

每个任务完成后，请确认以下事项：

```
□ 代码编译通过
□ 单元测试通过
□ GoDoc 注释完整
□ 无 lint 警告
□ 与现有代码兼容
□ 示例代码可运行
```

### B. 提交规范

```
feat(module): 添加新功能
fix(module): 修复 bug
refactor(module): 重构代码
docs(module): 更新文档
test(module): 添加测试
chore: 其他变更
```

### C. 版本规划

| 版本 | 内容 | 预计时间 |
|------|------|----------|
| v2.0.0-alpha.1 | 第一、二阶段完成 | +2 周 |
| v2.0.0-alpha.2 | 第三阶段完成 | +1 周 |
| v2.0.0-beta.1 | 第四阶段完成 | +2 周 |
| v2.0.0-rc.1 | 第五阶段完成 | +1 周 |
| v2.0.0 | 正式发布 | +2 周 |

---

*最后更新: 2024年*

