export type PiholeInitStatus = 'UNINITIALIZED' | 'ADDED' | 'SKIPPED';

export interface FullInitStatus {
  userCreated: boolean;
  piholeStatus: PiholeInitStatus;
}

export interface User {
  username: string;
  createdAt: string;
  updatedAt: string;
}

export interface HttpError extends Error {
  status?: number;
}
