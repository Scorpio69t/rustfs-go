# 实施任务清单

## 1. 准备工作
- [ ] 1.1 创建 `examples/s3/` 目录结构
- [ ] 1.2 分析和分类所有 71 个旧示例文件
- [ ] 1.3 确定迁移优先级（核心功能 → 高级功能）
- [ ] 1.4 准备示例模板和代码风格指南

## 2. 核心示例迁移（第一批：基础操作）
- [ ] 2.1 存储桶基础操作（10个文件）
  - [ ] makebucket.go - 创建存储桶
  - [ ] removebucket.go - 删除存储桶
  - [ ] listbuckets.go - 列出存储桶
  - [ ] bucketexists.go - 检查存储桶是否存在
  - [ ] list-directory-buckets.go - 列出目录存储桶
- [ ] 2.2 对象基础操作（15个文件）
  - [ ] putobject.go - 上传对象
  - [ ] getobject.go - 下载对象
  - [ ] fputobject.go - 从文件上传
  - [ ] fgetobject.go - 下载到文件
  - [ ] statobject.go - 获取对象信息
  - [ ] removeobject.go - 删除对象
  - [ ] removeobjects.go - 批量删除对象
  - [ ] copyobject.go - 复制对象
  - [ ] composeobject.go - 组合对象
- [ ] 2.3 对象列表操作（5个文件）
  - [ ] listobjects.go - 列出对象（V1）
  - [ ] listobjectsV2.go - 列出对象（V2）
  - [ ] listobjects-N.go - 带限制列出
  - [ ] listincompleteuploads.go - 列出未完成上传
  - [ ] removeincompleteupload.go - 删除未完成上传

## 3. 高级功能示例迁移（第二批：版本控制和标签）
- [ ] 3.1 版本控制（5个文件）
  - [ ] enableversioning.go - 启用版本控制
  - [ ] suspendversioning.go - 暂停版本控制
  - [ ] getbucketversioning.go - 获取版本控制状态
  - [ ] listobjectversions.go - 列出对象版本
- [ ] 3.2 对象标签（7个文件）
  - [ ] putobject-with-tags.go - 上传带标签对象
  - [ ] putobjecttagging.go - 设置对象标签
  - [ ] getobjecttagging.go - 获取对象标签
  - [ ] removeobjecttagging.go - 删除对象标签
  - [ ] copyobject-with-new-tags.go - 复制对象并设置新标签
  - [ ] putbuckettagging.go - 设置存储桶标签
  - [ ] getbuckettagging.go - 获取存储桶标签
  - [ ] removebuckettagging.go - 删除存储桶标签

## 4. 加密和安全示例迁移（第三批）
- [ ] 4.1 服务端加密（6个文件）
  - [ ] put-encrypted-object.go - 上传加密对象
  - [ ] get-encrypted-object.go - 下载加密对象
  - [ ] fputencrypted-object.go - 从文件上传加密
  - [ ] putobject-getobject-sse.go - SSE 示例
  - [ ] setbucketencryption.go - 设置存储桶加密
  - [ ] getbucketencryption.go - 获取存储桶加密
  - [ ] removebucketencryption.go - 删除存储桶加密
- [ ] 4.2 客户端加密（2个文件）
  - [ ] putobject-client-encryption.go - 客户端加密上传
  - [ ] getobject-client-encryption.go - 客户端加密下载
- [ ] 4.3 对象锁定和保留（5个文件）
  - [ ] setobjectlockconfig.go - 设置对象锁定配置
  - [ ] getobjectlockconfig.go - 获取对象锁定配置
  - [ ] putobjectlegalhold.go - 设置法律保留
  - [ ] getobjectlegalhold.go - 获取法律保留
  - [ ] putobjectretention.go - 设置对象保留
  - [ ] getobjectretention.go - 获取对象保留

## 5. 策略和配置示例迁移（第四批）
- [ ] 5.1 存储桶策略（2个文件）
  - [ ] setbucketpolicy.go - 设置存储桶策略
  - [ ] getbucketpolicy.go - 获取存储桶策略
- [ ] 5.2 生命周期管理（2个文件）
  - [ ] setbucketlifecycle.go - 设置生命周期规则
  - [ ] getbucketlifecycle.go - 获取生命周期规则
- [ ] 5.3 跨区复制（3个文件）
  - [ ] setbucketreplication.go - 设置复制配置
  - [ ] getbucketreplication.go - 获取复制配置
  - [ ] removebucketreplication.go - 删除复制配置
- [ ] 5.4 事件通知（2个文件）
  - [ ] setbucketnotification.go - 设置事件通知
  - [ ] getbucketnotification.go - 获取事件通知
  - [ ] removeallbucketnotification.go - 删除所有通知
- [ ] 5.5 CORS 配置（1个文件）
  - [ ] putbucketcors.go - 设置 CORS 配置
- [ ] 5.6 ACL 配置（2个文件）
  - [ ] getobjectacl.go - 获取对象 ACL

## 6. 预签名 URL 示例迁移（第五批）
- [ ] 6.1 预签名操作（4个文件）
  - [ ] presignedgetobject.go - 预签名 GET URL
  - [ ] presignedputobject.go - 预签名 PUT URL
  - [ ] presignedheadobject.go - 预签名 HEAD URL
  - [ ] presignedpostpolicy.go - 预签名 POST Policy
- [ ] 6.2 带响应头覆盖（1个文件）
  - [ ] getobject-override-respheaders.go - 覆盖响应头

## 7. 高级上传和特殊功能（第六批）
- [ ] 7.1 流式和进度（3个文件）
  - [ ] putobject-streaming.go - 流式上传
  - [ ] putobject-progress.go - 带进度上传
  - [ ] putobject-s3-accelerate.go - S3 加速上传
- [ ] 7.2 校验和（1个文件）
  - [ ] putobject-checksum.go - 带校验和上传
- [ ] 7.3 对象恢复和查询（3个文件）
  - [ ] restoreobject.go - 恢复对象
  - [ ] restoreobject-select.go - 恢复并查询
  - [ ] selectobject.go - 对象查询
- [ ] 7.4 健康检查（1个文件）
  - [ ] healthcheck.go - 健康检查

## 8. 代码质量和文档
- [ ] 8.1 为所有示例添加清晰的中文注释
- [ ] 8.2 统一错误处理模式
- [ ] 8.3 添加必要的使用说明（README）
- [ ] 8.4 确保所有示例可以独立运行
- [ ] 8.5 验证代码格式（gofmt）
- [ ] 8.6 检查是否有未使用的导入

## 9. 测试和验证
- [ ] 9.1 对所有迁移的示例进行编译测试
- [ ] 9.2 选择核心示例进行功能测试
- [ ] 9.3 验证新 API 覆盖了旧示例的所有功能
- [ ] 9.4 确认无版权问题（无 MinIO 声明）

## 10. 文档更新
- [ ] 10.1 更新 README.md 中的示例引用
- [ ] 10.2 更新 README.zh.md 中的示例引用
- [ ] 10.3 在 examples/s3/README.md 中添加示例索引
- [ ] 10.4 更新 CHANGELOG.md 记录此变更
