# api Specification Deltas

## ADDED Requirements

### Requirement: Presigned HEAD URL
The SDK SHALL provide an API to generate presigned HEAD URLs with optional signed headers and query parameters.

#### Scenario: Generate presigned HEAD URL
- **WHEN** a caller requests a presigned HEAD URL for an object
- **THEN** the SDK returns a URL containing a valid signature and expiration
- **AND** the caller can include optional signed headers

### Requirement: GetObject response header overrides
The SDK SHALL allow callers to specify response header overrides for GetObject via query parameters (e.g., response-content-type, response-content-disposition).

#### Scenario: Override response headers on GET
- **WHEN** a caller requests an object with response header overrides
- **THEN** the SDK encodes the overrides as query parameters
- **AND** the request is signed with those parameters

### Requirement: List multipart uploads and parts
The SDK SHALL expose APIs to list multipart uploads for a bucket and list parts for a specific upload, including pagination controls.

#### Scenario: List multipart uploads
- **WHEN** a caller requests multipart upload listing for a bucket
- **THEN** the SDK returns upload entries and pagination tokens

#### Scenario: List upload parts
- **WHEN** a caller requests parts for a multipart upload
- **THEN** the SDK returns parts and pagination tokens

### Requirement: Checksum mode support
The SDK SHALL allow callers to enable checksum mode for uploads and propagate checksum-related headers or query parameters.

#### Scenario: Enable checksum mode for upload
- **WHEN** a caller opts into checksum mode
- **THEN** the SDK sends the appropriate checksum mode header(s)

### Requirement: S3 Accelerate support
The SDK SHALL allow callers to enable S3 Accelerate endpoints for compatible object operations.

#### Scenario: Use accelerate endpoint
- **WHEN** accelerate is enabled in client or operation options
- **THEN** the SDK resolves the accelerate endpoint and sends requests to it

### Requirement: Client-side encryption (CSE)
The SDK SHALL provide helper APIs to encrypt and decrypt object data on the client side and integrate with Put/Get flows.

#### Scenario: Upload encrypted data with CSE
- **WHEN** a caller uploads with a client-side encryption helper
- **THEN** the SDK encrypts the payload and stores required metadata

#### Scenario: Download and decrypt with CSE
- **WHEN** a caller downloads an object encrypted with CSE
- **THEN** the SDK decrypts the payload using provided keys
