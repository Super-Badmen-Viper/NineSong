# 媒体库同步上传下载客户端实现指南

本文档详细介绍了如何在客户端实现完整的媒体库同步上传下载管理页面功能，包括界面设计、交互逻辑和API调用方式。

## 目录
1. [概述](#概述)
2. [页面布局设计](#页面布局设计)
3. [核心功能实现](#核心功能实现)
4. [API调用示例](#api调用示例)
5. [进度监控实现](#进度监控实现)
6. [错误处理策略](#错误处理策略)
7. [最佳实践](#最佳实践)

## 概述

媒体库同步上传下载管理功能允许用户上传音频文件到指定媒体库，同时支持查看上传/下载进度、管理已上传的文件等操作。客户端需要实现直观的用户界面和流畅的用户体验。

## 页面布局设计

### 主页面结构
```html
<div class="media-library-page">
  <!-- 顶部控制栏 -->
  <div class="control-panel">
    <button class="btn-upload">上传音频文件</button>
    <select class="library-selector">选择媒体库</select>
    <input type="text" class="search-input" placeholder="搜索文件...">
  </div>
  
  <!-- 进度监控面板 -->
  <div class="progress-panel">
    <h3>上传/下载进度</h3>
    <div class="progress-list"></div>
  </div>
  
  <!-- 文件列表 -->
  <div class="file-grid">
    <div class="file-item" v-for="file in files">
      <div class="file-info">
        <span class="file-name">{{ file.fileName }}</span>
        <span class="file-size">{{ formatFileSize(file.fileSize) }}</span>
      </div>
      <div class="file-actions">
        <button class="btn-download">下载</button>
        <button class="btn-delete">删除</button>
      </div>
    </div>
  </div>
</div>
```

### 功能区域划分
1. **控制区**: 上传按钮、媒体库选择、搜索功能
2. **进度区**: 实时显示上传/下载进度
3. **文件区**: 显示已上传的文件列表
4. **详情区**: 显示选中文件的详细信息

## 核心功能实现

### 1. 文件上传功能

#### HTML 表单
```html
<form id="upload-form" enctype="multipart/form-data">
  <input type="file" id="audio-file" accept="audio/*" multiple>
  <select id="library-select">
    <option value="">选择媒体库</option>
  </select>
  <button type="submit">开始上传</button>
</form>
```

#### JavaScript 实现
```javascript
class MediaLibraryManager {
  constructor(apiBaseUrl, accessToken) {
    this.apiBaseUrl = apiBaseUrl;
    this.accessToken = accessToken;
    this.uploadQueue = [];
    this.activeUploads = new Map();
  }

  // 上传单个文件
  async uploadFile(file, libraryId, uploaderId) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', uploaderId);

    try {
      const response = await fetch(`${this.apiBaseUrl}/media-library-audio/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.accessToken}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`上传失败: ${response.statusText}`);
      }

      const result = await response.json();
      return result.data.result; // 根据实际API响应结构调整
    } catch (error) {
      console.error('文件上传失败:', error);
      throw error;
    }
  }

  // 分块上传实现（用于大文件）
  async uploadFileInChunks(file, libraryId, uploaderId) {
    const chunkSize = 1024 * 1024; // 1MB chunks
    const totalChunks = Math.ceil(file.size / chunkSize);
    const uploadId = this.generateUploadId();

    for (let i = 0; i < totalChunks; i++) {
      const start = i * chunkSize;
      const end = Math.min(start + chunkSize, file.size);
      const chunk = file.slice(start, end);

      const formData = new FormData();
      formData.append('file', chunk);
      formData.append('library_id', libraryId);
      formData.append('uploader_id', uploaderId);
      formData.append('upload_id', uploadId);
      formData.append('chunk_index', i);
      formData.append('total_chunks', totalChunks);
      formData.append('is_last_chunk', i === totalChunks - 1);

      const response = await fetch(`${this.apiBaseUrl}/media-library-audio/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.accessToken}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`分块上传失败: ${response.statusText}`);
      }
    }
  }

  // 生成唯一的上传ID
  generateUploadId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }
}
```

### 2. 进度监控功能

#### 获取上传进度
```javascript
// 获取特定上传任务的进度
async function getUploadProgress(uploadId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/progress/upload/${uploadId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`获取进度失败: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.result;
  } catch (error) {
    console.error('获取上传进度失败:', error);
    throw error;
  }
}

// 获取下载进度
async function getDownloadProgress(fileId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/progress/download/${fileId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`获取进度失败: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.result;
  } catch (error) {
    console.error('获取下载进度失败:', error);
    throw error;
  }
}
```

#### 实时进度更新
```javascript
class ProgressTracker {
  constructor(updateCallback) {
    this.updateCallback = updateCallback;
    this.trackedTasks = new Set();
    this.pollingInterval = 1000; // 1秒轮询一次
  }

  // 开始跟踪进度
  startTracking(taskId, taskType) {
    this.trackedTasks.add({ id: taskId, type: taskType });
    this.startPolling();
  }

  // 停止跟踪进度
  stopTracking(taskId) {
    this.trackedTasks.delete(taskId);
    if (this.trackedTasks.size === 0) {
      this.stopPolling();
    }
  }

  // 开始轮询进度
  startPolling() {
    if (this.pollingActive) return;
    this.pollingActive = true;
    this.poll();
  }

  // 停止轮询
  stopPolling() {
    this.pollingActive = false;
    if (this.pollTimer) {
      clearTimeout(this.pollTimer);
    }
  }

  // 轮询函数
  async poll() {
    if (!this.pollingActive) return;

    for (const task of this.trackedTasks) {
      try {
        let progress;
        if (task.type === 'upload') {
          progress = await getUploadProgress(task.id);
        } else if (task.type === 'download') {
          progress = await getDownloadProgress(task.id);
        }

        this.updateCallback(task.id, progress);
      } catch (error) {
        console.error(`获取${task.type}进度失败:`, error);
      }
    }

    this.pollTimer = setTimeout(() => this.poll(), this.pollingInterval);
  }
}
```

### 3. 文件管理功能

#### 获取文件列表
```javascript
// 获取指定媒体库的文件列表
async function getFilesByLibrary(libraryId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/library/${libraryId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`获取文件列表失败: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.results;
  } catch (error) {
    console.error('获取文件列表失败:', error);
    throw error;
  }
}

// 获取指定用户上传的文件列表
async function getFilesByUploader(uploaderId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/uploader/${uploaderId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`获取用户文件列表失败: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.results;
  } catch (error) {
    console.error('获取用户文件列表失败:', error);
    throw error;
  }
}

// 删除文件
async function deleteFile(fileId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/${fileId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`删除文件失败: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error('删除文件失败:', error);
    throw error;
  }
}
```

## API调用示例

### 1. 上传音频文件
```javascript
// 上传单个文件
async function uploadAudioFile(fileInput, libraryId) {
  const file = fileInput.files[0];
  if (!file) {
    alert('请选择音频文件');
    return;
  }

  // 检查文件大小（例如限制为500MB）
  if (file.size > 500 * 1024 * 1024) {
    // 对大文件使用分块上传
    await mediaManager.uploadFileInChunks(file, libraryId, userId);
  } else {
    // 对小文件使用普通上传
    const result = await mediaManager.uploadFile(file, libraryId, userId);
    console.log('上传成功:', result);
  }
}
```

### 2. 下载音频文件
```javascript
// 下载音频文件
function downloadAudioFile(fileId, fileName) {
  const downloadUrl = `${apiBaseUrl}/media-library-audio/download/${fileId}`;
  
  // 创建隐藏的下载链接
  const link = document.createElement('a');
  link.href = downloadUrl;
  link.download = fileName || 'audio_file';
  link.style.display = 'none';
  
  // 添加认证头
  fetch(downloadUrl, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  })
  .then(response => response.blob())
  .then(blob => {
    const url = window.URL.createObjectURL(blob);
    link.href = url;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  })
  .catch(error => {
    console.error('下载失败:', error);
  });
}
```

### 3. 获取文件详细信息
```javascript
// 获取文件详细信息
async function getFileInfo(fileId) {
  try {
    const response = await fetch(`${apiBaseUrl}/media-library-audio/info/${fileId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`获取文件信息失败: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.result;
  } catch (error) {
    console.error('获取文件信息失败:', error);
    throw error;
  }
}
```

## 进度监控实现

### 1. 进度条组件
```html
<div class="progress-container">
  <div class="progress-bar">
    <div class="progress-fill" :style="{ width: progress + '%' }"></div>
  </div>
  <div class="progress-text">{{ Math.round(progress) }}%</div>
  <div class="progress-status">{{ status }}</div>
</div>
```

### 2. 进度更新逻辑
```javascript
// 更新进度显示
function updateProgressDisplay(taskId, progressData) {
  const progressElement = document.querySelector(`[data-task-id="${taskId}"]`);
  if (!progressElement) return;

  const progressBar = progressElement.querySelector('.progress-fill');
  const progressText = progressElement.querySelector('.progress-text');
  const progressStatus = progressElement.querySelector('.progress-status');

  if (progressBar) {
    progressBar.style.width = `${progressData.progress}%`;
  }
  
  if (progressText) {
    progressText.textContent = `${Math.round(progressData.progress)}%`;
  }
  
  if (progressStatus) {
    progressStatus.textContent = progressData.status;
  }
}
```

### 3. 批量上传处理
```javascript
// 批量上传多个文件
async function batchUpload(files, libraryId) {
  const uploadPromises = [];

  for (const file of files) {
    const promise = mediaManager.uploadFile(file, libraryId, userId)
      .then(result => {
        console.log(`文件 ${file.name} 上传成功:`, result);
        return { file: file.name, status: 'success', result };
      })
      .catch(error => {
        console.error(`文件 ${file.name} 上传失败:`, error);
        return { file: file.name, status: 'error', error: error.message };
      });
    
    uploadPromises.push(promise);
  }

  // 等待所有上传完成
  const results = await Promise.all(uploadPromises);
  return results;
}
```

## 错误处理策略

### 1. 网络错误处理
```javascript
// 通用请求函数，包含错误处理
async function apiRequest(url, options = {}) {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        ...options.headers
      }
    });

    if (!response.ok) {
      if (response.status === 401) {
        // 未授权，可能需要重新登录
        handleAuthError();
        throw new Error('认证失败，请重新登录');
      } else if (response.status >= 500) {
        // 服务器错误
        throw new Error('服务器错误，请稍后重试');
      } else {
        // 其他客户端错误
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `请求失败: ${response.statusText}`);
      }
    }

    return await response.json();
  } catch (error) {
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      // 网络连接问题
      throw new Error('网络连接失败，请检查网络连接');
    }
    throw error;
  }
}
```

### 2. 文件类型验证
```javascript
// 验证音频文件类型
function validateAudioFile(file) {
  const validTypes = ['audio/mpeg', 'audio/wav', 'audio/flac', 'audio/aac', 'audio/mp4', 'audio/ogg', 'audio/x-ms-wma'];
  const validExtensions = ['.mp3', '.wav', '.flac', '.aac', '.m4a', '.ogg', '.wma'];

  // 检查MIME类型
  if (!validTypes.includes(file.type)) {
    const fileExt = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
    if (!validExtensions.includes(fileExt)) {
      return false;
    }
  }

  return true;
}
```

## 最佳实践

### 1. 用户体验优化
- 提供清晰的上传/下载进度指示
- 实现断点续传功能（对于大文件）
- 支持拖拽上传
- 提供批量操作功能
- 实现预览功能（如音频波形图）

### 2. 性能优化
- 对大文件实现分块上传
- 使用压缩技术减少传输数据量
- 实现合理的缓存策略
- 优化API调用频率，避免不必要的请求

### 3. 安全考虑
- 验证文件类型和大小
- 实现适当的访问控制
- 对敏感信息进行加密传输
- 防止恶意文件上传

### 4. 响应式设计
```css
/* 响应式布局 */
@media (max-width: 768px) {
  .control-panel {
    flex-direction: column;
    gap: 10px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }
  
  .progress-panel {
    order: -1; /* 在移动设备上将进度面板移到上方 */
  }
}
```

## 总结

通过以上实现，客户端可以提供一个功能完整、用户体验良好的媒体库同步上传下载管理系统。关键是要关注进度监控、错误处理和性能优化，确保用户能够顺畅地管理和同步他们的音频文件。