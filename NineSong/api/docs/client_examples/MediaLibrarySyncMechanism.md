# 媒体库同步机制详解与最佳实践

本文档详细介绍了媒体库音频文件上传下载的同步机制，包括后台处理、状态管理、错误恢复等方面的最佳实践。

## 目录
1. [同步机制概述](#同步机制概述)
2. [同步记录模型详解](#同步记录模型详解)
3. [同步状态管理](#同步状态管理)
4. [后台任务处理](#后台任务处理)
5. [错误处理与恢复机制](#错误处理与恢复机制)
6. [性能优化策略](#性能优化策略)
7. [监控与日志](#监控与日志)
8. [安全考虑](#安全考虑)

## 同步机制概述

媒体库同步机制是确保音频文件在上传、下载过程中状态一致性的重要组件。该机制通过同步记录跟踪每个文件操作的生命周期，提供完整的审计轨迹和状态监控能力。

### 核心组件
1. **同步记录模型** - 记录每次同步操作的详细信息
2. **状态管理器** - 管理同步记录的生命周期状态
3. **后台处理器** - 异步处理长时间运行的同步任务
4. **进度追踪器** - 实时监控同步进度

### 同步流程
```
用户发起操作 → 创建同步记录 → 后台处理 → 更新状态 → 触发扫描 → 完成
```

## 同步记录模型详解

### 字段说明
```go
type MediaLibrarySyncRecord struct {
    ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    MediaFileID         primitive.ObjectID `bson:"media_file_id" json:"media_file_id"`                    // 关联的媒体文件ID
    MediaLibraryAudioID primitive.ObjectID `bson:"media_library_audio_id" json:"media_library_audio_id"` // 关联的媒体库音频文件ID
    SyncStatus          string             `bson:"sync_status" json:"sync_status"`                       // 同步状态
    SyncType            string             `bson:"sync_type" json:"sync_type"`                           // 同步类型: upload, download, both
    LastSyncTime        primitive.DateTime `bson:"last_sync_time" json:"last_sync_time"`                 // 最后同步时间
    CreatedAt           primitive.DateTime `bson:"created_at" json:"created_at"`
    UpdatedAt           primitive.DateTime `bson:"updated_at" json:"updated_at"`
    ErrorMessage        string             `bson:"error_message,omitempty" json:"error_message,omitempty"` // 错误信息
    Progress            float64            `bson:"progress" json:"progress"`                             // 同步进度 (0-100)
    FileSize            int64              `bson:"file_size" json:"file_size"`                           // 文件大小
    FilePath            string             `bson:"file_path" json:"file_path"`                           // 文件路径
    FileName            string             `bson:"file_name" json:"file_name"`                           // 文件名
    Checksum            string             `bson:"checksum" json:"checksum"`                             // 文件校验和
}
```

### 状态流转
- `pending` → `uploading`/`downloading` → `processing` → `synced`/`failed`
- 每个状态都有明确的进入和退出条件

### 索引策略
为确保查询性能，应在以下字段上创建索引：
- `media_file_id` - 单字段索引
- `media_library_audio_id` - 单字段索引
- `sync_status` - 单字段索引
- `sync_type` - 单字段索引
- `media_file_id + sync_status` - 复合索引

## 同步状态管理

### 状态枚举定义
```javascript
const SYNC_STATUS = {
  PENDING: 'pending',      // 待处理
  UPLOADING: 'uploading',  // 上传中
  DOWNLOADING: 'downloading', // 下载中
  PROCESSING: 'processing', // 处理中
  SYNCED: 'synced',       // 已同步
  FAILED: 'failed',       // 同步失败
  CANCELLED: 'cancelled'   // 已取消
};

const SYNC_TYPE = {
  UPLOAD: 'upload',
  DOWNLOAD: 'download',
  BOTH: 'both'
};
```

### 状态转换规则
```javascript
const stateTransitions = {
  [SYNC_STATUS.PENDING]: [SYNC_STATUS.UPLOADING, SYNC_STATUS.DOWNLOADING, SYNC_STATUS.CANCELLED],
  [SYNC_STATUS.UPLOADING]: [SYNC_STATUS.PROCESSING, SYNC_STATUS.FAILED],
  [SYNC_STATUS.DOWNLOADING]: [SYNC_STATUS.PROCESSING, SYNC_STATUS.FAILED],
  [SYNC_STATUS.PROCESSING]: [SYNC_STATUS.SYNCED, SYNC_STATUS.FAILED],
  [SYNC_STATUS.FAILED]: [SYNC_STATUS.PENDING], // 支持重试
  [SYNC_STATUS.CANCELLED]: [],
  [SYNC_STATUS.SYNCED]: []
};
```

### 状态管理服务
```javascript
class SyncStateManager {
  constructor(syncApi) {
    this.syncApi = syncApi;
  }

  // 更新同步状态
  async updateSyncStatus(recordId, newStatus, progress = 0, errorMessage = '') {
    const updateData = {
      sync_status: newStatus,
      progress: progress,
      updated_at: new Date().toISOString()
    };

    if (errorMessage) {
      updateData.error_message = errorMessage;
    }

    if ([SYNC_STATUS.SYNCED, SYNC_STATUS.FAILED].includes(newStatus)) {
      updateData.last_sync_time = new Date().toISOString();
    }

    return await this.syncApi.update(recordId, updateData);
  }

  // 验证状态转换是否合法
  isValidTransition(currentStatus, newStatus) {
    return stateTransitions[currentStatus]?.includes(newStatus) || false;
  }

  // 处理同步完成
  async handleSyncCompletion(recordId, result) {
    await this.updateSyncStatus(
      recordId, 
      SYNC_STATUS.SYNCED, 
      100, 
      '', 
      {
        file_path: result.filePath,
        file_size: result.fileSize,
        checksum: result.checksum
      }
    );
  }

  // 处理同步失败
  async handleSyncFailure(recordId, error) {
    await this.updateSyncStatus(
      recordId, 
      SYNC_STATUS.FAILED, 
      0, 
      error.message
    );
  }
}
```

## 后台任务处理

### 任务队列实现
```javascript
class BackgroundSyncProcessor {
  constructor(maxConcurrentTasks = 3) {
    this.maxConcurrentTasks = maxConcurrentTasks;
    this.runningTasks = new Set();
    this.pendingTasks = [];
    this.paused = false;
  }

  // 添加同步任务
  async addTask(syncRecord, operation) {
    const taskId = this.generateTaskId();
    const task = {
      id: taskId,
      syncRecord,
      operation,
      createdAt: new Date(),
      retries: 0,
      maxRetries: 3
    };

    this.pendingTasks.push(task);
    this.processNext();

    return taskId;
  }

  // 处理下一个任务
  async processNext() {
    if (this.paused || this.runningTasks.size >= this.maxConcurrentTasks || this.pendingTasks.length === 0) {
      return;
    }

    const task = this.pendingTasks.shift();
    this.runningTasks.add(task.id);

    try {
      await this.executeTask(task);
    } catch (error) {
      await this.handleTaskFailure(task, error);
    } finally {
      this.runningTasks.delete(task.id);
      // 继续处理下一个任务
      setTimeout(() => this.processNext(), 100);
    }
  }

  // 执行任务
  async executeTask(task) {
    const { syncRecord, operation } = task;
    
    // 更新状态为处理中
    await syncStateManager.updateSyncStatus(
      syncRecord.id, 
      operation.type === 'upload' ? SYNC_STATUS.UPLOADING : SYNC_STATUS.DOWNLOADING
    );

    try {
      // 执行具体操作
      const result = await operation.execute();
      
      // 更新进度
      await syncStateManager.updateSyncStatus(
        syncRecord.id, 
        SYNC_STATUS.PROCESSING, 
        100
      );

      // 标记为同步完成
      await syncStateManager.handleSyncCompletion(syncRecord.id, result);
    } catch (error) {
      throw error;
    }
  }

  // 处理任务失败
  async handleTaskFailure(task, error) {
    if (task.retries < task.maxRetries) {
      // 递增重试次数并重新加入队列
      task.retries++;
      this.pendingTasks.unshift(task); // 优先重试
      console.warn(`Task ${task.id} failed, retrying (${task.retries}/${task.maxRetries}):`, error.message);
    } else {
      // 达到最大重试次数，标记为失败
      await syncStateManager.handleSyncFailure(task.syncRecord.id, error);
      console.error(`Task ${task.id} permanently failed:`, error.message);
    }
  }

  // 生成任务ID
  generateTaskId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  // 暂停处理
  pause() {
    this.paused = true;
  }

  // 恢复处理
  resume() {
    this.paused = false;
    this.processNext();
  }
}
```

### 长时间运行任务的处理
```javascript
// Web Worker 实现长时间任务处理
// sync-worker.js
class SyncWorker {
  constructor() {
    this.workers = new Map();
  }

  // 创建音频处理工作线程
  createAudioProcessingWorker(audioFile, onProgress, onComplete, onError) {
    const worker = new Worker('/workers/audio-processor.js');
    const workerId = this.generateWorkerId();

    worker.onmessage = (e) => {
      const { type, data } = e.data;
      
      switch (type) {
        case 'progress':
          onProgress(data.progress, data.status);
          break;
        case 'complete':
          onComplete(data.result);
          this.cleanupWorker(workerId);
          break;
        case 'error':
          onError(data.error);
          this.cleanupWorker(workerId);
          break;
      }
    };

    worker.postMessage({
      type: 'processAudio',
      file: audioFile
    });

    this.workers.set(workerId, worker);
    return workerId;
  }

  cleanupWorker(workerId) {
    const worker = this.workers.get(workerId);
    if (worker) {
      worker.terminate();
      this.workers.delete(workerId);
    }
  }

  generateWorkerId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }
}
```

## 错误处理与恢复机制

### 错误分类
1. **网络错误** - 连接超时、断网等
2. **服务器错误** - 5xx错误、服务不可用
3. **客户端错误** - 4xx错误、参数错误
4. **文件错误** - 文件损坏、权限不足
5. **同步错误** - 状态不一致、冲突

### 错误处理策略
```javascript
class ErrorHandler {
  static errorHandlingRules = {
    // 网络错误 - 可重试
    NETWORK_ERROR: {
      retryable: true,
      backoffMultiplier: 2,
      maxDelay: 60000 // 1分钟
    },
    // 服务器错误 - 可重试
    SERVER_ERROR: {
      retryable: true,
      backoffMultiplier: 1.5,
      maxDelay: 30000
    },
    // 客户端错误 - 不重试
    CLIENT_ERROR: {
      retryable: false,
      backoffMultiplier: 1,
      maxDelay: 0
    },
    // 文件错误 - 不重试
    FILE_ERROR: {
      retryable: false,
      backoffMultiplier: 1,
      maxDelay: 0
    }
  };

  // 计算重试延迟时间
  static calculateRetryDelay(attempt, errorType) {
    const rule = this.errorHandlingRules[errorType];
    if (!rule.retryable) return 0;

    const baseDelay = Math.pow(rule.backoffMultiplier, attempt) * 1000; // 基础1秒
    return Math.min(baseDelay, rule.maxDelay);
  }

  // 处理错误
  static async handleError(error, context) {
    const errorType = this.classifyError(error);
    const rule = this.errorHandlingRules[errorType];

    if (rule.retryable && context.attempt < context.maxAttempts) {
      const delay = this.calculateRetryDelay(context.attempt, errorType);
      
      console.warn(`Attempt ${context.attempt + 1} failed, retrying in ${delay}ms:`, error.message);
      
      await this.delay(delay);
      return { shouldRetry: true, delay };
    } else {
      console.error('Permanent failure:', error.message);
      return { shouldRetry: false, error };
    }
  }

  // 错误分类
  static classifyError(error) {
    if (error.message.includes('NetworkError') || error.message.includes('timeout')) {
      return 'NETWORK_ERROR';
    } else if (error.status >= 500) {
      return 'SERVER_ERROR';
    } else if (error.status >= 400 && error.status < 500) {
      return 'CLIENT_ERROR';
    } else if (error.message.includes('file') || error.message.includes('permission')) {
      return 'FILE_ERROR';
    }
    
    return 'SERVER_ERROR'; // 默认归类为服务器错误
  }

  static delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### 断点续传实现
```javascript
class ResumeUploadManager {
  constructor() {
    this.uploadSessions = new Map(); // 存储上传会话
  }

  // 检查是否支持断点续传
  async checkResumeCapability(file, uploadId) {
    try {
      // 查询服务器上的上传会话
      const response = await fetch(`/media-library-audio/progress/upload/${uploadId}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        const progress = await response.json();
        return {
          canResume: true,
          uploadedBytes: progress.data.result.uploaded_bytes || 0,
          status: progress.data.result.status
        };
      }
    } catch (error) {
      console.warn('Could not check resume capability:', error.message);
    }

    return { canResume: false, uploadedBytes: 0, status: 'new' };
  }

  // 恢复上传
  async resumeUpload(file, uploadId, libraryId, uploaderId, startByte = 0) {
    const chunkSize = 1024 * 1024; // 1MB chunks
    const totalChunks = Math.ceil((file.size - startByte) / chunkSize);
    
    for (let i = 0; i < totalChunks; i++) {
      const absoluteStart = startByte + (i * chunkSize);
      const absoluteEnd = Math.min(absoluteStart + chunkSize, file.size);
      const chunk = file.slice(absoluteStart, absoluteEnd);

      const formData = new FormData();
      formData.append('file', chunk);
      formData.append('library_id', libraryId);
      formData.append('uploader_id', uploaderId);
      formData.append('upload_id', uploadId);
      formData.append('chunk_index', Math.floor(absoluteStart / chunkSize));
      formData.append('total_chunks', Math.ceil(file.size / chunkSize));
      formData.append('is_last_chunk', absoluteEnd === file.size);

      const response = await fetch('/media-library-audio/upload', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error(`Upload chunk failed: ${response.statusText}`);
      }

      // 更新进度
      const progress = ((absoluteEnd - startByte) / (file.size - startByte)) * 100;
      this.onProgress?.(progress);
    }

    return { uploadId, status: 'completed' };
  }
}
```

## 性能优化策略

### 1. 批量处理优化
```javascript
class BatchSyncOptimizer {
  constructor(batchSize = 10, concurrency = 3) {
    this.batchSize = batchSize;
    this.concurrency = concurrency;
  }

  // 批量同步处理
  async batchSync(syncRecords) {
    const batches = this.createBatches(syncRecords, this.batchSize);
    const results = [];

    for (const batch of batches) {
      const batchResults = await this.processBatch(batch);
      results.push(...batchResults);
      
      // 批次间延迟，避免服务器压力过大
      await this.delay(100);
    }

    return results;
  }

  // 创建批次
  createBatches(array, batchSize) {
    const batches = [];
    for (let i = 0; i < array.length; i += batchSize) {
      batches.push(array.slice(i, i + batchSize));
    }
    return batches;
  }

  // 处理单个批次
  async processBatch(batch) {
    const promises = batch.map(async (record) => {
      try {
        const result = await this.processSingleRecord(record);
        return { record, result, status: 'success' };
      } catch (error) {
        return { record, error, status: 'error' };
      }
    });

    // 限制并发数量
    const results = [];
    for (let i = 0; i < promises.length; i += this.concurrency) {
      const batchPromises = promises.slice(i, i + this.concurrency);
      const batchResults = await Promise.all(batchPromises);
      results.push(...batchResults);
    }

    return results;
  }

  async processSingleRecord(record) {
    // 具体的同步逻辑
    return await syncService.syncRecord(record);
  }

  async delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

### 2. 缓存策略
```javascript
class SyncCacheManager {
  constructor() {
    this.cache = new Map();
    this.maxCacheSize = 1000;
    this.ttl = 5 * 60 * 1000; // 5分钟
  }

  // 获取缓存
  get(key) {
    const cached = this.cache.get(key);
    if (cached && Date.now() - cached.timestamp < this.ttl) {
      return cached.value;
    } else {
      this.cache.delete(key);
      return null;
    }
  }

  // 设置缓存
  set(key, value) {
    // 检查缓存大小
    if (this.cache.size >= this.maxCacheSize) {
      // 删除最旧的项目
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }

    this.cache.set(key, {
      value,
      timestamp: Date.now()
    });
  }

  // 预热常用数据
  async warmUp(syncIds) {
    const uncachedIds = syncIds.filter(id => !this.get(`sync_${id}`));
    
    if (uncachedIds.length > 0) {
      const freshData = await syncService.getMultiple(uncachedIds);
      freshData.forEach(item => {
        this.set(`sync_${item.id}`, item);
      });
    }
  }
}
```

## 监控与日志

### 1. 日志记录
```javascript
class SyncLogger {
  constructor(level = 'info') {
    this.level = level;
    this.levels = {
      debug: 0,
      info: 1,
      warn: 2,
      error: 3
    };
  }

  log(level, message, meta = {}) {
    if (this.levels[level] >= this.levels[this.level]) {
      const logEntry = {
        timestamp: new Date().toISOString(),
        level,
        message,
        meta,
        correlationId: meta.correlationId || this.generateCorrelationId()
      };

      // 输出到控制台
      console[level === 'error' ? 'error' : level === 'warn' ? 'warn' : 'log'](logEntry);

      // 发送到远程日志服务（可选）
      this.sendToRemoteLog(logEntry);
    }
  }

  debug(message, meta) { this.log('debug', message, meta); }
  info(message, meta) { this.log('info', message, meta); }
  warn(message, meta) { this.log('warn', message, meta); }
  error(message, meta) { this.log('error', message, meta); }

  generateCorrelationId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  async sendToRemoteLog(logEntry) {
    // 实现发送到远程日志服务的逻辑
    try {
      await fetch('/logs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(logEntry)
      });
    } catch (error) {
      // 静默失败，不影响主流程
    }
  }
}

// 使用示例
const logger = new SyncLogger('info');

class SyncProcessor {
  constructor() {
    this.logger = new SyncLogger('info');
  }

  async processSync(syncRecord) {
    const correlationId = this.logger.generateCorrelationId();
    this.logger.info('Starting sync process', { 
      syncId: syncRecord.id, 
      correlationId 
    });

    try {
      const result = await this.executeSync(syncRecord);
      this.logger.info('Sync completed successfully', { 
        syncId: syncRecord.id, 
        correlationId,
        duration: result.duration
      });
      return result;
    } catch (error) {
      this.logger.error('Sync failed', { 
        syncId: syncRecord.id, 
        correlationId,
        error: error.message
      });
      throw error;
    }
  }
}
```

### 2. 性能监控
```javascript
class SyncPerformanceMonitor {
  constructor() {
    this.metrics = {
      syncDuration: [], // 同步耗时分布
      throughput: [],   // 吞吐量
      errorRate: [],    // 错误率
      concurrentTasks: 0 // 并发任务数
    };
  }

  // 记录同步指标
  recordSyncMetric(duration, fileSize, success = true) {
    this.metrics.syncDuration.push({
      duration,
      fileSize,
      success,
      timestamp: Date.now()
    });

    // 保持最近1000条记录
    if (this.metrics.syncDuration.length > 1000) {
      this.metrics.syncDuration = this.metrics.syncDuration.slice(-1000);
    }
  }

  // 计算平均同步速度
  getAverageSpeed() {
    const recentSyncs = this.metrics.syncDuration.filter(
      sync => sync.success && Date.now() - sync.timestamp < 5 * 60 * 1000 // 5分钟内
    );

    if (recentSyncs.length === 0) return 0;

    const totalSize = recentSyncs.reduce((sum, sync) => sum + sync.fileSize, 0);
    const totalTime = recentSyncs.reduce((sum, sync) => sum + sync.duration, 0);

    return totalTime > 0 ? totalSize / totalTime : 0; // bytes/ms
  }

  // 获取错误率
  getErrorRate() {
    const recentSyncs = this.metrics.syncDuration.filter(
      sync => Date.now() - sync.timestamp < 5 * 60 * 1000
    );

    if (recentSyncs.length === 0) return 0;

    const failedSyncs = recentSyncs.filter(sync => !sync.success);
    return failedSyncs.length / recentSyncs.length;
  }

  // 生成性能报告
  generateReport() {
    return {
      averageSyncTime: this.calculateAverageSyncTime(),
      averageSpeed: this.getAverageSpeed(),
      errorRate: this.getErrorRate(),
      activeTasks: this.metrics.concurrentTasks,
      totalSyncs: this.metrics.syncDuration.length
    };
  }

  calculateAverageSyncTime() {
    const recentSyncs = this.metrics.syncDuration.filter(
      sync => Date.now() - sync.timestamp < 5 * 60 * 1000
    );

    if (recentSyncs.length === 0) return 0;

    const totalDuration = recentSyncs.reduce((sum, sync) => sum + sync.duration, 0);
    return totalDuration / recentSyncs.length;
  }
}
```

## 安全考虑

### 1. 认证与授权
```javascript
class SecureSyncManager {
  constructor() {
    this.tokenExpiryBuffer = 5 * 60 * 1000; // 5分钟缓冲期
  }

  // 检查认证令牌是否有效
  isTokenValid(token) {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const expiryTime = payload.exp * 1000;
      return expiryTime - Date.now() > this.tokenExpiryBuffer;
    } catch (error) {
      return false;
    }
  }

  // 自动刷新令牌
  async refreshTokenIfNeeded() {
    const token = localStorage.getItem('token');
    if (!this.isTokenValid(token)) {
      try {
        const refreshToken = localStorage.getItem('refreshToken');
        const response = await fetch('/auth/refresh', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ refreshToken })
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('token', data.accessToken);
          localStorage.setItem('refreshToken', data.refreshToken);
          return data.accessToken;
        }
      } catch (error) {
        console.error('Token refresh failed:', error);
        throw new Error('Authentication expired');
      }
    }

    return localStorage.getItem('token');
  }

  // 添加认证头
  async addAuthHeaders(headers = {}) {
    const token = await this.refreshTokenIfNeeded();
    return {
      ...headers,
      'Authorization': `Bearer ${token}`
    };
  }
}
```

### 2. 文件安全验证
```javascript
class FileSecurityValidator {
  constructor() {
    this.allowedMimeTypes = [
      'audio/mpeg',
      'audio/wav', 
      'audio/flac',
      'audio/aac',
      'audio/mp4',
      'audio/ogg',
      'audio/x-ms-wma'
    ];
    
    this.maxFileSize = 500 * 1024 * 1024; // 500MB
    this.allowedExtensions = ['.mp3', '.wav', '.flac', '.aac', '.m4a', '.ogg', '.wma'];
  }

  // 验证文件安全性
  async validateFileSecurity(file) {
    const errors = [];

    // 检查文件大小
    if (file.size > this.maxFileSize) {
      errors.push(`文件大小超过限制 (${this.formatFileSize(this.maxFileSize)})`);
    }

    // 检查MIME类型
    if (!this.allowedMimeTypes.includes(file.type)) {
      errors.push(`不支持的文件类型: ${file.type}`);
    }

    // 检查文件扩展名
    const fileExt = file.name.toLowerCase().substring(file.name.lastIndexOf('.'));
    if (!this.allowedExtensions.includes(fileExt)) {
      errors.push(`不支持的文件扩展名: ${fileExt}`);
    }

    // 检查文件魔数（如果可能）
    const magicNumber = await this.getFileMagicNumber(file);
    if (!this.isValidAudioMagicNumber(magicNumber)) {
      errors.push('文件魔数验证失败，可能存在安全风险');
    }

    return {
      isValid: errors.length === 0,
      errors
    };
  }

  // 获取文件魔数
  async getFileMagicNumber(file) {
    const buffer = await file.slice(0, 16).arrayBuffer();
    const view = new Uint8Array(buffer);
    return Array.from(view.slice(0, 8)).map(b => b.toString(16).padStart(2, '0')).join('');
  }

  // 验证音频文件魔数
  isValidAudioMagicNumber(magicNumber) {
    // 简化的魔数验证，实际应用中需要更完整的魔数表
    const validSignatures = [
      'fffb', // MP3
      '52494646', // WAV
      '664c6143', // FLAC
      '494433'    // MP3 ID3 tag
    ];
    
    return validSignatures.some(sig => magicNumber.startsWith(sig));
  }

  formatFileSize(bytes) {
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Byte';
    const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  }
}
```

## 总结

媒体库同步机制是确保音频文件上传下载功能可靠性的核心组件。通过实施上述最佳实践，可以构建一个高效、安全、可靠的同步系统：

1. **状态管理**: 清晰的状态定义和流转规则确保操作的可追踪性
2. **错误处理**: 完善的错误分类和恢复机制提高系统的容错能力
3. **性能优化**: 批量处理和缓存策略提升系统吞吐量
4. **监控日志**: 详细的监控和日志便于问题诊断和性能分析
5. **安全保障**: 严格的认证授权和文件验证保护系统安全

这些实践可以根据具体业务需求进行调整和定制，以满足不同规模和复杂度的应用场景。