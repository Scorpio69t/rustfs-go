/*
 * RustFS Go SDK
 * Copyright 2025 RustFS Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package signer 提供 AWS Signature Version 4 和 Version 2 的签名实现。
//
// 这个包提供了用于 S3 兼容存储服务的 HTTP 请求签名功能。
//
// 使用示例：
//
//	req, _ := http.NewRequest("GET", "https://s3.amazonaws.com/bucket/key", nil)
//	signedReq := signer.SignV4(req, "access-key", "secret-key", "", "us-east-1", "s3")
package signer
