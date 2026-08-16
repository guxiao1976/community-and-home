// File Service types

export interface FileInfo {
  id: string;
  userId: string;
  entityType: string;
  entityId: string;
  fileName: string;
  filePath: string;
  fileSize: number;
  mimeType: string;
  bucketName: string;
  uploadTime: number;
  file_type: string;   // 白名单规范类型（ConfirmUpload magic-bytes 层产出）
  confirmed: boolean;  // 上传流程完成标志
}
