import { apiV1Fetch } from './client';
import type { GetConfigResponse, PatchConfigRequest, PatchConfigResponse } from '@/types/config';

export async function getConfig(): Promise<GetConfigResponse> {
	return apiV1Fetch<GetConfigResponse>('/config/');
}

export async function patchConfig(patch: PatchConfigRequest): Promise<PatchConfigResponse> {
	return apiV1Fetch<PatchConfigResponse>('/config/', {
		method: 'PATCH',
		body: JSON.stringify(patch),
	});
}
