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
}
