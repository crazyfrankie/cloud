# 文件预览与下载统一解决方案

## 问题背景

在云存储系统中，不同类型的文件在MinIO中的预览支持情况不同：
- **支持预览的文件**：图片、PDF、视频、音频等可以直接在浏览器中预览
- **不支持预览的文件**：doc、docx、压缩包等文件，访问预览URL时会直接下载

这导致了用户体验不一致的问题：有些文件点击是预览，有些是下载。

## 解决方案

### 1. 统一的文件操作API

#### 后端API设计

```
GET /api/files/{fileId}/preview  - 统一预览接口
GET /api/files/{fileId}/download - 统一下载接口
GET /api/files/{fileId}/text     - 文本文件预览接口
```

#### 文件类型分类

- **可预览文件** (preview)：图片、PDF、视频、音频等
- **文本文件** (text)：txt、md、json、代码文件等
- **下载文件** (download)：doc、docx、压缩包等

### 2. 智能文件操作决策

根据文件扩展名自动决定操作类型：

```go
// 配置文件类型支持情况
type FilePreviewConfig struct {
    PreviewableTypes map[string]bool  // 支持在线预览
    TextTypes        map[string]bool  // 文本文件
    ThumbnailTypes   map[string]bool  // 支持缩略图
}
```

### 3. 前端用户界面

#### 文件列表界面
- **文件名点击**：智能预览（使用 `/preview` 接口）
- **下载按钮**：明确下载（使用 `/download` 接口）
- **操作统一**：所有文件都有相同的交互方式

#### 预览行为
- **图片/PDF/视频**：直接在浏览器预览
- **文本文件**：重定向到专门的文本预览页面
- **Office文档**：直接下载

### 4. 实现细节

#### 后端实现

1. **文件操作信息增强**
```go
type FileResp struct {
    // ...existing fields...
    Action       string `json:"action"`       // preview/download/text
    PreviewURL   string `json:"previewUrl"`   // 预览URL
    DownloadURL  string `json:"downloadUrl"`  // 下载URL
    Previewable  bool   `json:"previewable"`  // 是否可预览
    HasThumbnail bool   `json:"hasThumbnail"` // 是否有缩略图
    ContentType  string `json:"contentType"`  // MIME类型
}
```

2. **统一预览接口**
```go
func (h *FileHandler) PreviewFile() gin.HandlerFunc {
    // 根据文件类型智能决定行为：
    // - 可预览文件：重定向到MinIO URL
    // - 文本文件：重定向到文本预览页面
    // - 其他文件：重定向到下载接口
}
```

#### 前端实现

1. **文件点击处理**
```typescript
const handleFileClick = (file: any) => {
  if (file.type === 'folder') {
    emit('navigate', file)
  } else {
    // 使用统一的预览API
    const previewUrl = `/api/files/${file.id}/preview`
    window.open(previewUrl, '_blank')
  }
}
```

2. **下载处理**
```typescript
const handleDownload = (file: any) => {
  const downloadUrl = `/api/files/${file.id}/download`
  // 强制下载，设置正确的文件名
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = file.name
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
```

### 5. 优势

#### 用户体验统一
- 所有文件都有相同的交互方式
- 预览和下载行为明确区分
- 文本文件有专门的预览界面

#### 技术架构清晰
- 后端统一处理文件操作逻辑
- 前端只需调用统一接口
- 易于扩展新的文件类型支持

#### 性能优化
- 文件类型判断在后端进行
- 避免前端重复逻辑
- 支持缓存和CDN优化

### 6. 文件类型支持列表

#### 可预览文件
- **图片**：jpg, jpeg, png, gif, bmp, webp, svg
- **文档**：pdf
- **视频**：mp4, webm, ogg
- **音频**：mp3, wav, ogg, aac

#### 文本文件（专门预览页面）
- **文本**：txt, md, json, xml, csv
- **网页**：html, htm, css
- **代码**：js, ts, go, py, java, cpp, c
- **配置**：yaml, yml, ini, conf, log

#### 下载文件
- **Office文档**：doc, docx, xls, xlsx, ppt, pptx
- **压缩包**：zip, rar, 7z
- **其他二进制文件**

### 7. 扩展性

可以轻松添加新的文件类型支持：
1. 在 `FilePreviewConfig` 中添加文件扩展名
2. 根据需要调整预览逻辑
3. 前端无需修改，自动支持新类型

这个解决方案提供了一个优雅、统一且可扩展的文件预览和下载体验。
