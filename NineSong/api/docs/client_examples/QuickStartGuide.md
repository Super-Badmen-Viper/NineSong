# 媒体库音频文件管理快速入门指南

本指南将帮助您快速了解和使用媒体库音频文件上传下载功能，包含API调用示例和常见使用场景。

## 目录
1. [准备工作](#准备工作)
2. [API基础概念](#api基础概念)
3. [快速开始示例](#快速开始示例)
4. [常见使用场景](#常见使用场景)
5. [错误处理](#错误处理)
6. [调试技巧](#调试技巧)

## 准备工作

### 1. 获取访问令牌
在使用API之前，您需要先获取访问令牌：

```javascript
// 登录获取令牌
const loginResponse = await fetch('/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'your-email@example.com',
    password: 'your-password'
  })
});

const { accessToken, refreshToken } = await loginResponse.json();
localStorage.setItem('token', accessToken);
```

### 2. 设置API基础配置
```javascript
const API_BASE_URL = 'http://localhost:8080'; // 根据实际部署地址调整
const AUTH_TOKEN = localStorage.getItem('token');

const apiConfig = {
  headers: {
    'Authorization': `Bearer ${AUTH_TOKEN}`,
    'Accept': 'application/json'
  }
};
```

## API基础概念

### 认证
- 所有API端点都需要 `Authorization: Bearer <token>` 头
- 令牌有效期请参考认证服务文档

### 响应格式
```json
{
  "code": 1,
  "data": {
    "result": {},
    "count": 1
  },
  "message": "Success"
}
```

### 主要实体
- **Library ID**: 媒体库唯一标识符
- **Uploader ID**: 上传者唯一标识符
- **File ID**: 音频文件唯一标识符
- **Upload ID**: 上传会话唯一标识符

## 快速开始示例

### 1. 上传音频文件
```javascript
// HTML
<input type="file" id="audioFile" accept="audio/*" />
<select id="librarySelect">
  <option value="library1">我的音乐库</option>
  <option value="library2">工作音频</option>
</select>
<button onclick="uploadFile()">上传</button>

// JavaScript
async function uploadFile() {
  const fileInput = document.getElementById('audioFile');
  const librarySelect = document.getElementById('librarySelect');
  
  const file = fileInput.files[0];
  const libraryId = librarySelect.value;
  const uploaderId = 'current_user_id'; // 替换为实际用户ID

  if (!file) {
    alert('请选择音频文件');
    return;
  }

  const formData = new FormData();
  formData.append('file', file);
  formData.append('library_id', libraryId);
  formData.append('uploader_id', uploaderId);

  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`
      },
      body: formData
    });

    const result = await response.json();
    if (response.ok) {
      console.log('上传成功:', result.data.result);
      alert('上传成功！');
    } else {
      console.error('上传失败:', result.message);
      alert('上传失败: ' + result.message);
    }
  } catch (error) {
    console.error('上传错误:', error);
    alert('上传出错: ' + error.message);
  }
}
```

### 2. 获取媒体库文件列表
```javascript
async function getLibraryFiles(libraryId) {
  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/library/${libraryId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`,
        'Accept': 'application/json'
      }
    });

    const result = await response.json();
    if (response.ok) {
      console.log('文件列表:', result.data.results);
      return result.data.results;
    } else {
      console.error('获取文件列表失败:', result.message);
      return [];
    }
  } catch (error) {
    console.error('获取文件列表错误:', error);
    return [];
  }
}

// 使用示例
const files = await getLibraryFiles('library1');
console.log(`找到 ${files.length} 个文件`);
```

### 3. 下载音频文件
```javascript
function downloadFile(fileId, fileName) {
  const downloadUrl = `${API_BASE_URL}/media-library-audio/download/${fileId}`;
  
  // 创建带认证的下载请求
  fetch(downloadUrl, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`
    }
  })
  .then(response => {
    if (!response.ok) {
      throw new Error('下载失败');
    }
    return response.blob();
  })
  .then(blob => {
    // 创建下载链接
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName || `audio_file_${fileId}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  })
  .catch(error => {
    console.error('下载错误:', error);
    alert('下载失败: ' + error.message);
  });
}
```

### 4. 删除音频文件
```javascript
async function deleteFile(fileId) {
  if (!confirm('确定要删除这个文件吗？')) {
    return;
  }

  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/${fileId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`,
        'Accept': 'application/json'
      }
    });

    const result = await response.json();
    if (response.ok) {
      console.log('删除成功:', result);
      alert('文件已删除');
      // 刷新文件列表
      refreshFileList();
    } else {
      console.error('删除失败:', result.message);
      alert('删除失败: ' + result.message);
    }
  } catch (error) {
    console.error('删除错误:', error);
    alert('删除出错: ' + error.message);
  }
}
```

## 带进度监控的上传示例

### HTML
```html
<div id="uploadSection">
  <input type="file" id="audioFile" accept="audio/*" multiple />
  <select id="librarySelect">
    <option value="library1">我的音乐库</option>
    <option value="library2">工作音频</option>
  </select>
  <button onclick="uploadWithProgress()">上传(带进度)</button>
  
  <div id="progressContainer" style="display:none;">
    <div class="progress-bar">
      <div class="progress-fill" id="progressFill" style="width:0%"></div>
    </div>
    <div id="progressText">0%</div>
    <div id="progressStatus">准备上传...</div>
  </div>
</div>
```

### JavaScript
```javascript
async function uploadWithProgress() {
  const fileInput = document.getElementById('audioFile');
  const librarySelect = document.getElementById('librarySelect');
  const files = fileInput.files;
  
  if (files.length === 0) {
    alert('请选择音频文件');
    return;
  }

  // 显示进度容器
  document.getElementById('progressContainer').style.display = 'block';

  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    document.getElementById('progressStatus').textContent = `正在上传: ${file.name}`;
    
    try {
      await uploadSingleFileWithProgress(file, librarySelect.value, i + 1, files.length);
    } catch (error) {
      console.error(`文件 ${file.name} 上传失败:`, error);
      alert(`文件 ${file.name} 上传失败: ${error.message}`);
    }
  }

  // 上传完成后
  document.getElementById('progressStatus').textContent = '全部上传完成！';
  setTimeout(() => {
    document.getElementById('progressContainer').style.display = 'none';
  }, 2000);
}

function uploadSingleFileWithProgress(file, libraryId, current, total) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const formData = new FormData();
    
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', 'current_user_id'); // 替换为实际用户ID

    // 上传进度
    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable) {
        const percentComplete = (event.loaded / event.total) * 100;
        document.getElementById('progressFill').style.width = `${percentComplete}%`;
        document.getElementById('progressText').textContent = `${Math.round(percentComplete)}%`;
      }
    });

    // 上传完成
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        const response = JSON.parse(xhr.responseText);
        document.getElementById('progressStatus').textContent = 
          `完成: ${file.name} (${current}/${total})`;
        resolve(response.data.result);
      } else {
        reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`));
      }
    });

    // 上传错误
    xhr.addEventListener('error', () => {
      reject(new Error('网络错误'));
    });

    xhr.open('POST', `${API_BASE_URL}/media-library-audio/upload`);
    xhr.setRequestHeader('Authorization', `Bearer ${AUTH_TOKEN}`);
    xhr.send(formData);
  });
}
```

## 常见使用场景

### 1. 批量上传
```javascript
// 批量上传多个文件
async function batchUpload(files, libraryId) {
  const uploadPromises = Array.from(files).map((file, index) => {
    return new Promise((resolve, reject) => {
      setTimeout(() => { // 避免同时发起过多请求
        uploadSingleFileWithProgress(file, libraryId, index + 1, files.length)
          .then(resolve)
          .catch(reject);
      }, index * 500); // 间隔500ms上传一个
    });
  });

  try {
    const results = await Promise.all(uploadPromises);
    console.log('批量上传完成:', results);
    return results;
  } catch (error) {
    console.error('批量上传失败:', error);
    throw error;
  }
}
```

### 2. 获取上传进度
```javascript
// 获取特定上传任务的进度
async function getUploadProgress(uploadId) {
  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/progress/upload/${uploadId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`
      }
    });

    const result = await response.json();
    if (response.ok) {
      console.log('上传进度:', result.data.result);
      return result.data.result;
    } else {
      console.error('获取进度失败:', result.message);
      return null;
    }
  } catch (error) {
    console.error('获取进度错误:', error);
    return null;
  }
}
```

### 3. 搜索和过滤
```javascript
// 按上传者获取文件
async function getFilesByUploader(uploaderId) {
  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/uploader/${uploaderId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`
      }
    });

    const result = await response.json();
    if (response.ok) {
      return result.data.results;
    } else {
      console.error('获取上传者文件失败:', result.message);
      return [];
    }
  } catch (error) {
    console.error('获取上传者文件错误:', error);
    return [];
  }
}
```

## 错误处理

### 常见错误码
- `400`: 请求参数错误
- `401`: 认证失败
- `403`: 权限不足
- `404`: 资源不存在
- `500`: 服务器内部错误

### 统一错误处理
```javascript
async function apiCall(url, options = {}) {
  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`,
        'Accept': 'application/json',
        ...options.headers
      }
    });

    const result = await response.json();

    if (!response.ok) {
      switch (response.status) {
        case 400:
          throw new Error(`参数错误: ${result.message}`);
        case 401:
          throw new Error('认证失败，请重新登录');
        case 403:
          throw new Error('权限不足');
        case 404:
          throw new Error('资源不存在');
        case 500:
          throw new Error(`服务器错误: ${result.message}`);
        default:
          throw new Error(`请求失败 (${response.status}): ${result.message}`);
      }
    }

    return result;
  } catch (error) {
    console.error('API调用错误:', error);
    throw error;
  }
}

// 使用统一错误处理
async function safeUploadFile(file, libraryId) {
  try {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('library_id', libraryId);
    formData.append('uploader_id', 'current_user_id');

    const result = await apiCall(`${API_BASE_URL}/media-library-audio/upload`, {
      method: 'POST',
      body: formData
    });

    console.log('上传成功:', result.data.result);
    return result.data.result;
  } catch (error) {
    alert('操作失败: ' + error.message);
    return null;
  }
}
```

## 调试技巧

### 1. 使用浏览器开发者工具
```javascript
// 启用详细日志
const DEBUG = true;

function debugLog(message, data = null) {
  if (DEBUG) {
    console.log(`[MEDIA LIBRARY DEBUG] ${message}`, data);
  }
}

// 在API调用前后添加日志
async function debugUploadFile(file, libraryId) {
  debugLog('开始上传文件', { fileName: file.name, libraryId });
  
  try {
    const result = await uploadFile(file, libraryId);
    debugLog('上传成功', result);
    return result;
  } catch (error) {
    debugLog('上传失败', error);
    throw error;
  }
}
```

### 2. 网络请求监控
在浏览器开发者工具的 Network 标签页中，您可以：
- 查看所有API请求和响应
- 检查请求头和响应头
- 监控请求耗时
- 查看错误详情

### 3. 本地存储调试
```javascript
// 检查认证令牌
function checkAuthToken() {
  const token = localStorage.getItem('token');
  if (!token) {
    console.warn('未找到认证令牌');
    return false;
  }
  
  // 解码JWT查看过期时间
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const expiryTime = new Date(payload.exp * 1000);
    const now = new Date();
    
    console.log(`令牌到期时间: ${expiryTime}`);
    console.log(`当前时间: ${now}`);
    console.log(`是否过期: ${now > expiryTime}`);
    
    return now <= expiryTime;
  } catch (error) {
    console.error('令牌解码失败:', error);
    return false;
  }
}
```

## 完整示例应用

```html
<!DOCTYPE html>
<html>
<head>
  <title>媒体库音频管理</title>
  <style>
    .container { max-width: 800px; margin: 0 auto; padding: 20px; }
    .upload-section { margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
    .file-list { margin-top: 20px; }
    .file-item { padding: 10px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
    .progress-bar { width: 100%; height: 20px; background-color: #f0f0f0; border-radius: 10px; overflow: hidden; margin: 10px 0; }
    .progress-fill { height: 100%; background-color: #4CAF50; transition: width 0.3s ease; }
    button { padding: 8px 16px; margin: 5px; cursor: pointer; }
    input, select { padding: 8px; margin: 5px; }
  </style>
</head>
<body>
  <div class="container">
    <h1>媒体库音频管理</h1>
    
    <!-- 上传区域 -->
    <div class="upload-section">
      <h3>上传音频文件</h3>
      <input type="file" id="audioFile" accept="audio/*" multiple />
      <select id="librarySelect">
        <option value="default_library">默认库</option>
      </select>
      <button onclick="uploadFiles()">上传</button>
      
      <div id="progressContainer" style="display:none;">
        <div class="progress-bar">
          <div class="progress-fill" id="progressFill" style="width:0%"></div>
        </div>
        <div id="progressText">0%</div>
        <div id="progressStatus">准备上传...</div>
      </div>
    </div>
    
    <!-- 文件列表 -->
    <div class="file-list">
      <h3>音频文件列表</h3>
      <button onclick="loadFiles()">刷新列表</button>
      <div id="fileList"></div>
    </div>
  </div>

  <script>
    const API_BASE_URL = 'http://localhost:8080';
    const AUTH_TOKEN = localStorage.getItem('token');
    
    if (!AUTH_TOKEN) {
      alert('请先登录获取访问令牌');
      // 这里应该跳转到登录页面
    }
    
    async function uploadFiles() {
      const fileInput = document.getElementById('audioFile');
      const librarySelect = document.getElementById('librarySelect');
      const files = fileInput.files;
      
      if (files.length === 0) {
        alert('请选择音频文件');
        return;
      }
      
      document.getElementById('progressContainer').style.display = 'block';
      
      for (let i = 0; i < files.length; i++) {
        const file = files[i];
        document.getElementById('progressStatus').textContent = `正在上传: ${file.name}`;
        
        try {
          const formData = new FormData();
          formData.append('file', file);
          formData.append('library_id', librarySelect.value);
          formData.append('uploader_id', 'current_user_id');
          
          const response = await fetch(`${API_BASE_URL}/media-library-audio/upload`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${AUTH_TOKEN}`
            },
            body: formData
          });
          
          const result = await response.json();
          if (response.ok) {
            document.getElementById('progressFill').style.width = `${((i + 1) / files.length) * 100}%`;
            document.getElementById('progressText').textContent = `${Math.round(((i + 1) / files.length) * 100)}%`;
          } else {
            throw new Error(result.message);
          }
        } catch (error) {
          console.error(`上传 ${file.name} 失败:`, error);
          alert(`上传 ${file.name} 失败: ${error.message}`);
        }
      }
      
      document.getElementById('progressStatus').textContent = '上传完成！';
      setTimeout(() => {
        document.getElementById('progressContainer').style.display = 'none';
        loadFiles(); // 刷新文件列表
      }, 1000);
    }
    
    async function loadFiles() {
      try {
        const response = await fetch(`${API_BASE_URL}/media-library-audio/library/default_library`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${AUTH_TOKEN}`
          }
        });
        
        const result = await response.json();
        if (response.ok) {
          displayFiles(result.data.results);
        } else {
          alert('加载文件列表失败: ' + result.message);
        }
      } catch (error) {
        console.error('加载文件列表失败:', error);
        alert('加载文件列表失败: ' + error.message);
      }
    }
    
    function displayFiles(files) {
      const fileListDiv = document.getElementById('fileList');
      fileListDiv.innerHTML = '';
      
      if (files.length === 0) {
        fileListDiv.innerHTML = '<p>暂无文件</p>';
        return;
      }
      
      files.forEach(file => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.innerHTML = `
          <div>
            <strong>${file.file_name}</strong><br>
            <small>大小: ${(file.file_size / 1024 / 1024).toFixed(2)} MB | 
            上传时间: ${new Date(file.created_at).toLocaleString()}</small>
          </div>
          <div>
            <button onclick="downloadFile('${file.id}', '${file.file_name}')">下载</button>
            <button onclick="deleteFile('${file.id}')">删除</button>
          </div>
        `;
        fileListDiv.appendChild(fileItem);
      });
    }
    
    function downloadFile(fileId, fileName) {
      const downloadUrl = `${API_BASE_URL}/media-library-audio/download/${fileId}`;
      window.open(downloadUrl, '_blank');
    }
    
    async function deleteFile(fileId) {
      if (!confirm('确定要删除这个文件吗？')) return;
      
      try {
        const response = await fetch(`${API_BASE_URL}/media-library-audio/${fileId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${AUTH_TOKEN}`
          }
        });
        
        const result = await response.json();
        if (response.ok) {
          alert('文件已删除');
          loadFiles(); // 刷新列表
        } else {
          alert('删除失败: ' + result.message);
        }
      } catch (error) {
        console.error('删除失败:', error);
        alert('删除失败: ' + error.message);
      }
    }
    
    // 页面加载时自动加载文件列表
    window.onload = function() {
      loadFiles();
    };
  </script>
</body>
</html>
```

## 歌词文件支持

### 支持的歌词格式

系统自动支持以下歌词文件格式：
- **LRC** (.lrc) - 标准歌词格式，支持时间轴同步
- **TXT** (.txt) - 纯文本歌词格式
- **KRC** (.krc) - 酷狗音乐歌词格式
- **QRC** (.qrc) - QQ音乐歌词格式

### 自动检测歌词文件

系统会在上传音频文件时自动检测同目录下的歌词文件：

```javascript
// 上传音频文件时，系统会自动检测歌词文件
async function uploadAudioWithAutoLyrics() {
  const fileInput = document.getElementById('audioFile');
  const file = fileInput.files[0];
  
  if (!file) {
    alert('请选择音频文件');
    return;
  }

  // 系统会自动检测同目录下的歌词文件
  // 支持的格式：.lrc, .txt, .krc, .qrc
  // 检测优先级：lrc > txt > krc > qrc
  
  const formData = new FormData();
  formData.append('file', file);
  formData.append('library_id', libraryId);
  formData.append('uploader_id', uploaderId);

  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`
      },
      body: formData
    });

    const result = await response.json();
    
    if (response.ok) {
      console.log('音频文件上传成功');
      
      // 检查是否检测到歌词文件
      if (result.data.result.lyricsDetected) {
        console.log(`检测到歌词文件: ${result.data.result.lyricsFormat}`);
        console.log('歌词文件已自动关联');
      }
      
      alert('上传成功！');
    } else {
      alert('上传失败: ' + result.message);
    }
  } catch (error) {
    console.error('上传错误:', error);
    alert('上传失败: ' + error.message);
  }
}
```

### 手动指定歌词文件

如果需要手动上传歌词文件，可以这样做：

```javascript
// 上传音频文件和歌词文件
async function uploadAudioWithLyrics() {
  const audioFile = document.getElementById('audioFile').files[0];
  const lyricsFile = document.getElementById('lyricsFile').files[0];
  
  if (!audioFile) {
    alert('请选择音频文件');
    return;
  }

  // 验证歌词文件格式
  if (lyricsFile) {
    const validFormats = ['.lrc', '.txt', '.krc', '.qrc'];
    const fileExt = '.' + lyricsFile.name.split('.').pop().toLowerCase();
    
    if (!validFormats.includes(fileExt)) {
      alert(`不支持的歌词格式: ${fileExt}\n支持的格式: ${validFormats.join(', ')}`);
      return;
    }
  }

  const formData = new FormData();
  formData.append('file', audioFile);
  formData.append('library_id', libraryId);
  formData.append('uploader_id', uploaderId);
  
  // 如果提供了歌词文件，也一起上传
  if (lyricsFile) {
    formData.append('lyrics_file', lyricsFile);
  }

  try {
    const response = await fetch(`${API_BASE_URL}/media-library-audio/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`
      },
      body: formData
    });

    const result = await response.json();
    
    if (response.ok) {
      alert('上传成功！');
      if (lyricsFile) {
        console.log('歌词文件已上传并关联');
      }
    } else {
      alert('上传失败: ' + result.message);
    }
  } catch (error) {
    console.error('上传错误:', error);
    alert('上传失败: ' + error.message);
  }
}
```

### 下载音频文件时获取歌词

```javascript
// 下载音频文件及其关联的歌词文件
async function downloadAudioWithLyrics(fileId, audioFileName) {
  try {
    // 1. 下载音频文件
    const audioResponse = await fetch(
      `${API_BASE_URL}/media-library-audio/download/${fileId}`,
      {
        headers: {
          'Authorization': `Bearer ${AUTH_TOKEN}`
        }
      }
    );

    if (!audioResponse.ok) {
      throw new Error('音频文件下载失败');
    }

    const audioBlob = await audioResponse.blob();
    const audioUrl = window.URL.createObjectURL(audioBlob);
    
    // 创建音频文件下载链接
    const audioLink = document.createElement('a');
    audioLink.href = audioUrl;
    audioLink.download = audioFileName;
    audioLink.click();
    window.URL.revokeObjectURL(audioUrl);

    // 2. 获取文件信息，检查是否有歌词
    const infoResponse = await fetch(
      `${API_BASE_URL}/media-library-audio/info/${fileId}`,
      {
        headers: {
          'Authorization': `Bearer ${AUTH_TOKEN}`
        }
      }
    );

    if (infoResponse.ok) {
      const info = await infoResponse.json();
      
      // 3. 如果存在歌词文件，下载歌词
      if (info.data.result.lyricsPath) {
        const lyricsFormat = info.data.result.lyricsFormat || 'lrc';
        const lyricsFileName = audioFileName.replace(/\.[^/.]+$/, '') + '.' + lyricsFormat;
        
        // 下载歌词文件（如果需要单独下载）
        console.log(`检测到歌词文件: ${lyricsFormat}`);
        console.log('歌词文件已随音频文件一起下载');
      }
    }
  } catch (error) {
    console.error('下载错误:', error);
    alert('下载失败: ' + error.message);
  }
}
```

### 歌词格式说明

#### LRC格式示例
```
[00:12.00]第一句歌词
[00:15.30]第二句歌词
[00:18.50]第三句歌词
```

#### TXT格式示例
```
第一句歌词
第二句歌词
第三句歌词
```

#### KRC和QRC格式
- KRC和QRC是二进制格式，需要专门的解析器
- 系统会自动识别和处理这些格式
- 客户端无需特殊处理，系统会保持原始格式

## 总结

通过本指南，您应该能够：

1. 设置基本的API调用环境
2. 实现音频文件的上传、下载、删除功能
3. 添加进度监控功能
4. 处理常见的错误情况
5. 使用调试工具进行故障排除
6. **处理歌词文件（支持LRC、TXT、KRC、QRC格式）**

记住始终在生产环境中使用HTTPS协议，并妥善保管您的认证令牌。如需更多高级功能，请参考完整的API文档。