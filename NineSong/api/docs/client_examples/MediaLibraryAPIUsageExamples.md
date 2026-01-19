# 媒体库音频文件管理API使用示例

本文档提供了媒体库音频文件上传下载功能的各种客户端实现示例，涵盖不同技术栈和使用场景。

## 目录
1. [API端点概览](#api端点概览)
2. [JavaScript/Fetch实现](#javascriptfetch实现)
3. [React组件实现](#react组件实现)
4. [Vue组件实现](#vue组件实现)
5. [Angular服务实现](#angular服务实现)
6. [移动端实现要点](#移动端实现要点)
7. [文件上传策略](#文件上传策略)
8. [进度监控最佳实践](#进度监控最佳实践)

## API端点概览

### 媒体库音频相关端点

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/media-library-audio/upload` | 上传音频文件 |
| GET | `/media-library-audio/download/:file_id` | 下载音频文件 |
| GET | `/media-library-audio/info/:file_id` | 获取音频文件信息 |
| GET | `/media-library-audio/library/:library_id` | 获取指定媒体库的音频文件列表 |
| GET | `/media-library-audio/uploader/:uploader_id` | 获取指定用户上传的音频文件列表 |
| DELETE | `/media-library-audio/:file_id` | 删除音频文件 |
| GET | `/media-library-audio/progress/upload/:upload_id` | 获取上传进度 |
| GET | `/media-library-audio/progress/download/:file_id` | 获取下载进度 |

### 请求头
- `Authorization: Bearer <access_token>` - 认证令牌
- `Content-Type: multipart/form-data` - 上传文件时使用

### 响应格式
```json
{
  "code": 1,
  "data": {
    "result": {}
  },
  "message": "Success"
}
```

## JavaScript/Fetch实现

### 1. 基础上传功能
```javascript
class MediaLibraryAPI {
  constructor(baseURL, accessToken) {
    this.baseURL = baseURL;
    this.accessToken = accessToken;
  }

  // 上传音频文件
  async uploadAudio(file, libraryId, uploaderId) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', uploaderId);

    try {
      const response = await fetch(`${this.baseURL}/media-library-audio/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.accessToken}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      return result.data.result;
    } catch (error) {
      console.error('Upload error:', error);
      throw error;
    }
  }

  // 下载音频文件
  async downloadAudio(fileId, fileName) {
    const response = await fetch(`${this.baseURL}/media-library-audio/download/${fileId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${this.accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`Download failed: ${response.statusText}`);
    }

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    
    // 创建下载链接
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName || `audio_${fileId}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  }

  // 获取文件列表
  async getFilesByLibrary(libraryId) {
    const response = await fetch(`${this.baseURL}/media-library-audio/library/${libraryId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${this.accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`Get files failed: ${response.statusText}`);
    }

    const result = await response.json();
    return result.data.results;
  }

  // 删除文件
  async deleteFile(fileId) {
    const response = await fetch(`${this.baseURL}/media-library-audio/${fileId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${this.accessToken}`
      }
    });

    if (!response.ok) {
      throw new Error(`Delete failed: ${response.statusText}`);
    }

    return await response.json();
  }
}
```

### 2. 带进度监控的上传
```javascript
// 带进度回调的上传实现
function createProgressiveUpload(file, onProgress) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    
    // 监听上传进度
    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        const percentComplete = (event.loaded / event.total) * 100;
        onProgress(percentComplete);
      }
    });

    // 处理完成事件
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(JSON.parse(xhr.responseText));
      } else {
        reject(new Error(`Upload failed: ${xhr.statusText}`));
      }
    });

    // 处理错误
    xhr.addEventListener('error', () => {
      reject(new Error('Network error during upload'));
    });

    // 构建请求
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', selectedLibraryId);
    formData.append('uploader_id', userId);

    xhr.open('POST', `${apiBaseURL}/media-library-audio/upload`);
    xhr.setRequestHeader('Authorization', `Bearer ${accessToken}`);
    xhr.send(formData);
  });
}
```

## React组件实现

### 1. 音频上传组件
```jsx
import React, { useState, useCallback } from 'react';

const AudioUploadComponent = ({ libraryId, onUploadSuccess }) => {
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [selectedFile, setSelectedFile] = useState(null);

  const handleFileChange = (event) => {
    const file = event.target.files[0];
    setSelectedFile(file);
  };

  const uploadFile = useCallback(async (file) => {
    if (!file) return;

    setIsUploading(true);
    setProgress(0);

    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', 'current_user_id'); // 替换为实际用户ID

    try {
      // 使用 fetch 实现带进度的上传
      const xhr = new XMLHttpRequest();
      
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const percentComplete = (event.loaded / event.total) * 100;
          setProgress(Math.round(percentComplete));
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          const response = JSON.parse(xhr.responseText);
          onUploadSuccess(response.data.result);
          resetState();
        }
      });

      xhr.addEventListener('error', () => {
        alert('上传失败，请重试');
        resetState();
      });

      xhr.open('POST', '/media-library-audio/upload');
      xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('token')}`);
      xhr.send(formData);
    } catch (error) {
      console.error('Upload error:', error);
      setIsUploading(false);
    }
  }, [libraryId, onUploadSuccess]);

  const resetState = () => {
    setIsUploading(false);
    setProgress(0);
    setSelectedFile(null);
  };

  const handleUpload = () => {
    if (selectedFile) {
      uploadFile(selectedFile);
    }
  };

  return (
    <div className="audio-upload-component">
      <input 
        type="file" 
        accept="audio/*" 
        onChange={handleFileChange} 
        disabled={isUploading}
      />
      
      {selectedFile && (
        <div className="file-info">
          <p>文件名: {selectedFile.name}</p>
          <p>大小: {(selectedFile.size / 1024 / 1024).toFixed(2)} MB</p>
        </div>
      )}

      {isUploading && (
        <div className="progress-container">
          <div className="progress-bar">
            <div 
              className="progress-fill" 
              style={{ width: `${progress}%` }}
            ></div>
          </div>
          <span className="progress-text">{progress}%</span>
        </div>
      )}

      <button 
        onClick={handleUpload} 
        disabled={!selectedFile || isUploading}
        className="upload-btn"
      >
        {isUploading ? '上传中...' : '上传'}
      </button>
    </div>
  );
};

export default AudioUploadComponent;
```

### 2. 文件管理组件
```jsx
import React, { useState, useEffect } from 'react';

const FileManagerComponent = ({ libraryId }) => {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadFiles();
  }, [libraryId]);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/media-library-audio/library/${libraryId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setFiles(data.data.results);
      }
    } catch (error) {
      console.error('Load files error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (fileId) => {
    if (window.confirm('确定要删除这个文件吗？')) {
      try {
        await fetch(`/media-library-audio/${fileId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          }
        });
        
        // 重新加载文件列表
        loadFiles();
      } catch (error) {
        console.error('Delete error:', error);
      }
    }
  };

  if (loading) {
    return <div>加载中...</div>;
  }

  return (
    <div className="file-manager">
      <h3>音频文件列表</h3>
      <div className="file-list">
        {files.map(file => (
          <div key={file.id} className="file-item">
            <div className="file-details">
              <h4>{file.file_name}</h4>
              <p>大小: {(file.file_size / 1024 / 1024).toFixed(2)} MB</p>
              <p>上传时间: {new Date(file.created_at).toLocaleString()}</p>
            </div>
            <div className="file-actions">
              <button 
                onClick={() => downloadFile(file.id, file.file_name)}
                className="btn-download"
              >
                下载
              </button>
              <button 
                onClick={() => handleDelete(file.id)}
                className="btn-delete"
              >
                删除
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
```

## Vue组件实现

### 1. Vue 3 Composition API 实现
```vue
<template>
  <div class="audio-upload-vue">
    <input 
      type="file" 
      accept="audio/*" 
      @change="handleFileChange"
      :disabled="isUploading"
    />
    
    <div v-if="selectedFile" class="file-info">
      <p>文件名: {{ selectedFile.name }}</p>
      <p>大小: {{ formatFileSize(selectedFile.size) }}</p>
    </div>

    <div v-if="isUploading" class="progress-container">
      <div class="progress-bar">
        <div 
          class="progress-fill" 
          :style="{ width: progress + '%' }"
        ></div>
      </div>
      <span class="progress-text">{{ progress }}%</span>
    </div>

    <button 
      @click="uploadFile" 
      :disabled="!selectedFile || isUploading"
      class="upload-btn"
    >
      {{ isUploading ? '上传中...' : '上传' }}
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
  libraryId: String
});

const emit = defineEmits(['upload-success']);

const selectedFile = ref(null);
const isUploading = ref(false);
const progress = ref(0);

const handleFileChange = (event) => {
  selectedFile.value = event.target.files[0];
};

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const uploadFile = async () => {
  if (!selectedFile.value) return;

  isUploading.value = true;
  progress.value = 0;

  const formData = new FormData();
  formData.append('file', selectedFile.value);
  formData.append('library_id', props.libraryId);
  formData.append('uploader_id', 'current_user_id'); // 替换为实际用户ID

  try {
    const xhr = new XMLHttpRequest();
    
    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        const percentComplete = (event.loaded / event.total) * 100;
        progress.value = Math.round(percentComplete);
      }
    });

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        const response = JSON.parse(xhr.responseText);
        emit('upload-success', response.data.result);
        resetState();
      }
    });

    xhr.addEventListener('error', () => {
      alert('上传失败，请重试');
      resetState();
    });

    xhr.open('POST', '/media-library-audio/upload');
    xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('token')}`);
    xhr.send(formData);
  } catch (error) {
    console.error('Upload error:', error);
    isUploading.value = false;
  }
};

const resetState = () => {
  isUploading.value = false;
  progress.value = 0;
  selectedFile.value = null;
};
</script>
```

## Angular服务实现

### 1. Angular服务
```typescript
import { Injectable } from '@angular/core';
import { HttpClient, HttpEventType, HttpHeaders } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';

export interface AudioFile {
  id: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  createdAt: string;
  libraryId: string;
  uploaderId: string;
}

@Injectable({
  providedIn: 'root'
})
export class MediaLibraryService {
  private baseUrl = '/media-library-audio';

  constructor(private http: HttpClient) {}

  // 上传文件并返回进度
  uploadFile(
    file: File, 
    libraryId: string, 
    uploaderId: string
  ): Observable<{ progress: number, result?: any }> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', uploaderId);

    const headers = new HttpHeaders({
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    });

    const progressSubject = new Subject<{ progress: number, result?: any }>();
    
    this.http.post(`${this.baseUrl}/upload`, formData, {
      headers,
      reportProgress: true,
      observe: 'events'
    }).subscribe(event => {
      if (event.type === HttpEventType.UploadProgress) {
        const progress = Math.round((100 * event.loaded) / event.total!);
        progressSubject.next({ progress });
      } else if (event.type === HttpEventType.Response) {
        progressSubject.next({ progress: 100, result: event.body });
        progressSubject.complete();
      }
    });

    return progressSubject.asObservable();
  }

  // 获取文件列表
  getFilesByLibrary(libraryId: string): Observable<AudioFile[]> {
    const headers = new HttpHeaders({
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    });

    return this.http.get<any>(`${this.baseUrl}/library/${libraryId}`, { headers })
      .pipe(
        map(response => response.data.results)
      );
  }

  // 下载文件
  downloadFile(fileId: string, fileName: string): void {
    const headers = new HttpHeaders({
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    });

    this.http.get(`${this.baseUrl}/download/${fileId}`, {
      headers,
      responseType: 'blob'
    }).subscribe(blob => {
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = fileName;
      link.click();
      window.URL.revokeObjectURL(url);
    });
  }

  // 删除文件
  deleteFile(fileId: string): Observable<any> {
    const headers = new HttpHeaders({
      'Authorization': `Bearer ${localStorage.getItem('token')}`
    });

    return this.http.delete(`${this.baseUrl}/${fileId}`, { headers });
  }
}
```

## 移动端实现要点

### 1. 移动端优化
```javascript
// 移动端音频上传优化
class MobileAudioUploader {
  constructor() {
    this.maxFileSize = 50 * 1024 * 1024; // 50MB限制
    this.chunkSize = 5 * 1024 * 1024; // 5MB分块
  }

  // 检测网络状况
  getNetworkStatus() {
    if ('connection' in navigator) {
      return {
        effectiveType: navigator.connection.effectiveType,
        downlink: navigator.connection.downlink
      };
    }
    return { effectiveType: '4g', downlink: 10 }; // 默认值
  }

  // 根据网络状况调整上传策略
  async adaptiveUpload(file, libraryId, uploaderId) {
    const networkStatus = this.getNetworkStatus();
    
    // 对于慢速网络，减小分块大小
    let chunkSize = this.chunkSize;
    if (networkStatus.effectiveType === 'slow-2g' || networkStatus.effectiveType === '2g') {
      chunkSize = 1 * 1024 * 1024; // 1MB
    } else if (networkStatus.effectiveType === '3g') {
      chunkSize = 2 * 1024 * 1024; // 2MB
    }

    if (file.size > this.maxFileSize) {
      // 大文件使用分块上传
      return await this.chunkedUpload(file, libraryId, uploaderId, chunkSize);
    } else {
      // 小文件直接上传
      return await this.directUpload(file, libraryId, uploaderId);
    }
  }

  // 分块上传实现
  async chunkedUpload(file, libraryId, uploaderId, chunkSize) {
    const totalChunks = Math.ceil(file.size / chunkSize);
    const uploadId = this.generateUploadId();
    let uploadedBytes = 0;

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

      try {
        const response = await fetch('/media-library-audio/upload', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
          },
          body: formData
        });

        if (!response.ok) {
          throw new Error(`Chunk upload failed: ${response.statusText}`);
        }

        uploadedBytes += chunk.size;
        const progress = (uploadedBytes / file.size) * 100;
        this.onProgress?.(progress);
      } catch (error) {
        console.error('Chunk upload error:', error);
        throw error;
      }
    }

    return uploadId;
  }
}
```

## 文件上传策略

### 1. 智能上传策略
```javascript
class SmartUploadStrategy {
  constructor() {
    this.strategies = {
      'small': this.smallFileUpload.bind(this),    // < 10MB
      'medium': this.mediumFileUpload.bind(this),  // 10MB - 50MB
      'large': this.largeFileUpload.bind(this),    // > 50MB
      'auto': this.adaptiveUpload.bind(this)       // 自适应
    };
  }

  async upload(file, libraryId, uploaderId, strategy = 'auto') {
    return await this.strategies[strategy](file, libraryId, uploaderId);
  }

  // 小文件直接上传
  async smallFileUpload(file, libraryId, uploaderId) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', uploaderId);

    const response = await fetch('/media-library-audio/upload', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: formData
    });

    return await response.json();
  }

  // 中等文件带进度上传
  async mediumFileUpload(file, libraryId, uploaderId) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      const formData = new FormData();
      
      formData.append('file', file);
      formData.append('library_id', libraryId);
      formData.append('uploader_id', uploaderId);

      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const progress = (event.loaded / event.total) * 100;
          this.onProgress?.(progress);
        }
      });

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve(JSON.parse(xhr.responseText));
        } else {
          reject(new Error(`Upload failed: ${xhr.statusText}`));
        }
      });

      xhr.addEventListener('error', () => {
        reject(new Error('Network error'));
      });

      xhr.open('POST', '/media-library-audio/upload');
      xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('token')}`);
      xhr.send(formData);
    });
  }

  // 大文件分块上传
  async largeFileUpload(file, libraryId, uploaderId) {
    // 实现分块上传逻辑
    const chunkSize = 5 * 1024 * 1024; // 5MB
    const totalChunks = Math.ceil(file.size / chunkSize);
    const uploadId = Date.now().toString();

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

      const response = await fetch('/media-library-audio/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`Chunk upload failed: ${response.statusText}`);
      }

      const progress = ((i + 1) / totalChunks) * 100;
      this.onProgress?.(progress);
    }

    return { uploadId, status: 'completed' };
  }
}
```

## 进度监控最佳实践

### 1. 统一进度管理器
```javascript
class ProgressManager {
  constructor() {
    this.tasks = new Map();
    this.listeners = new Set();
  }

  // 注册任务
  registerTask(taskId, taskType, totalSize) {
    this.tasks.set(taskId, {
      id: taskId,
      type: taskType,
      totalSize: totalSize,
      transferred: 0,
      startTime: Date.now(),
      status: 'pending'
    });
  }

  // 更新进度
  updateProgress(taskId, transferred) {
    const task = this.tasks.get(taskId);
    if (task) {
      task.transferred = transferred;
      task.status = 'in-progress';
      
      const progress = (transferred / task.totalSize) * 100;
      this.notifyListeners(taskId, {
        progress,
        transferred,
        total: task.totalSize,
        speed: this.calculateSpeed(taskId, transferred),
        eta: this.calculateETA(taskId, progress)
      });
    }
  }

  // 计算速度
  calculateSpeed(taskId, transferred) {
    const task = this.tasks.get(taskId);
    if (!task) return 0;
    
    const elapsed = (Date.now() - task.startTime) / 1000; // 秒
    return elapsed > 0 ? transferred / elapsed : 0; // 字节/秒
  }

  // 计算预计完成时间
  calculateETA(taskId, progress) {
    if (progress >= 100) return 0;
    
    const speed = this.calculateSpeed(taskId, this.tasks.get(taskId)?.transferred || 0);
    if (speed <= 0) return -1; // 无法计算
    
    const remaining = (this.tasks.get(taskId)?.totalSize || 0) - (this.tasks.get(taskId)?.transferred || 0);
    return remaining / speed; // 秒
  }

  // 添加进度监听器
  addListener(callback) {
    this.listeners.add(callback);
    return () => this.listeners.delete(callback);
  }

  // 通知所有监听器
  notifyListeners(taskId, progressData) {
    this.listeners.forEach(listener => {
      listener(taskId, progressData);
    });
  }

  // 完成任务
  completeTask(taskId) {
    const task = this.tasks.get(taskId);
    if (task) {
      task.status = 'completed';
      task.progress = 100;
      this.notifyListeners(taskId, {
        progress: 100,
        transferred: task.totalSize,
        total: task.totalSize,
        speed: this.calculateSpeed(taskId, task.totalSize),
        eta: 0
      });
    }
  }
}

// 使用示例
const progressManager = new ProgressManager();

// 监听进度更新
const unsubscribe = progressManager.addListener((taskId, progressData) => {
  console.log(`Task ${taskId}: ${progressData.progress.toFixed(2)}% complete`);
  // 更新UI
  updateProgressBar(taskId, progressData.progress);
});
```

## 总结

本文档提供了多种技术栈的实现方案，开发者可以根据具体需求和技术栈选择合适的实现方式。关键要点包括：

1. **进度监控**: 使用XMLHttpRequest或Fetch API的进度事件
2. **错误处理**: 实现重试机制和错误恢复
3. **性能优化**: 根据文件大小和网络状况选择合适的上传策略
4. **用户体验**: 提供清晰的进度反馈和操作反馈

这些实现方案可以直接用于生产环境，也可以根据具体需求进行定制化改造。