# api Specification Deltas

## ADDED Requirements

### Requirement: Server-Side Encryption Support
RustFS Go SDK SHALL 支持 AWS S3 服务端加密（SSE），包括 SSE-S3、SSE-C 和 SSE-KMS 三种模式。

#### Scenario: Upload with SSE-S3
- **GIVEN** a RustFS client is configured
- **WHEN** user uploads an object with SSE-S3 encryption
- **THEN** the object is encrypted server-side
- **AND** the encryption header is set in the request

#### Scenario: Upload with SSE-C
- **GIVEN** a RustFS client is configured
- **AND** user provides a 256-bit encryption key
- **WHEN** user uploads an object with SSE-C encryption
- **THEN** the object is encrypted with the provided key
- **AND** the key is not stored on the server

---

### Requirement: CORS Configuration
RustFS Go SDK SHALL 支持跨域资源共享（CORS）配置，允许浏览器端直接访问 S3 资源。

#### Scenario: Set CORS configuration
- **GIVEN** a bucket exists
- **WHEN** user sets CORS rules allowing GET from example.com
- **THEN** CORS configuration is saved
- **AND** browsers can make cross-origin requests

---

### Requirement: Object Locking and Retention
RustFS Go SDK SHALL 支持对象锁定、法律保留和对象保留期，满足合规性要求。

#### Scenario: Enable object locking on bucket
- **GIVEN** a new bucket is created with object locking enabled
- **WHEN** an object is uploaded
- **THEN** the object can be protected from deletion

#### Scenario: Set legal hold
- **GIVEN** an object exists in a locked bucket
- **WHEN** user sets legal hold to ON
- **THEN** the object cannot be deleted or modified
- **AND** legal hold remains until explicitly removed

---

### Requirement: Access Control Lists (ACL)
RustFS Go SDK SHALL 支持细粒度的访问控制列表（ACL）管理。

#### Scenario: Get object ACL
- **GIVEN** an object exists
- **WHEN** user retrieves object ACL
- **THEN** current permissions are returned

---

### Requirement: Bucket Replication
RustFS Go SDK SHALL 支持存储桶跨区域/跨账户复制配置。

#### Scenario: Configure replication
- **GIVEN** source and destination buckets exist
- **WHEN** user configures replication rules
- **THEN** new objects are replicated automatically

---

### Requirement: Event Notification
RustFS Go SDK SHALL 支持 S3 事件通知配置和监听。

#### Scenario: Set notification configuration
- **GIVEN** a bucket exists
- **WHEN** user configures notification for PUT events
- **THEN** notifications are sent when objects are created

---

### Requirement: Object Composition
RustFS Go SDK SHALL 支持将多个对象组合成一个新对象。

#### Scenario: Compose multiple objects
- **GIVEN** three source objects exist
- **WHEN** user composes them into a new object
- **THEN** a single object containing all source data is created

---

### Requirement: Object Append Extension
RustFS Go SDK SHALL 支持对象追加操作（RustFS 扩展功能）。

#### Scenario: Append data to object
- **GIVEN** an object exists with size 100 bytes
- **WHEN** user appends 50 bytes at offset 100
- **THEN** object size becomes 150 bytes

---

### Requirement: Select Object Content
RustFS Go SDK SHALL 支持使用 SQL 表达式查询对象内容。

#### Scenario: Query CSV file
- **GIVEN** a CSV object exists
- **WHEN** user executes SELECT query with condition
- **THEN** only matching rows are returned

---

### Requirement: Restore Archived Objects
RustFS Go SDK SHALL 支持从归档存储（Glacier）恢复对象。

#### Scenario: Restore from Glacier
- **GIVEN** an object is in Glacier storage class
- **WHEN** user initiates restore for 7 days
- **THEN** object is restored and accessible for 7 days

---

### Requirement: Presigned POST Policy
RustFS Go SDK SHALL 支持生成浏览器直传的 POST Policy。

#### Scenario: Generate POST policy
- **GIVEN** a client configures POST policy
- **WHEN** policy specifies bucket, key, and expiration
- **THEN** a signed POST form is returned

---

### Requirement: Bucket Tagging
RustFS Go SDK SHALL 支持存储桶级别的标签。

#### Scenario: Set bucket tags
- **GIVEN** a bucket exists
- **WHEN** user sets tags Environment=Production
- **THEN** tags are saved and retrievable
