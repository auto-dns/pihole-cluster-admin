import { apiV1Fetch } from './client';
import type { AuditListResponse, RollbackResponse } from '@/types/audit';

export async function listAuditEntries(limit = 50, offset = 0): Promise<AuditListResponse> {
	return apiV1Fetch<AuditListResponse>(`/audit?limit=${limit}&offset=${offset}`);
}

export async function rollbackAuditEntry(id: number): Promise<RollbackResponse> {
	return apiV1Fetch<RollbackResponse>(`/audit/${id}/rollback`, { method: 'POST' });
}
