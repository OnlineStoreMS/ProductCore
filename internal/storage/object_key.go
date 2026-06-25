package storage

// BuildObjectKey 生成存储对象相对路径（local 与 minio 使用同一套规则）
func BuildObjectKey(rootPrefix string, opts UploadOptions, originalName, contentType string) string {
	filename := BuildStoredFilename(opts.Resource, safeFilename(originalName), contentType)
	return BuildRelativePath(rootPrefix, opts, filename)
}
