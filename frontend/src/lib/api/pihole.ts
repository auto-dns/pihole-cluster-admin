import apiFetch from './client';
import { PiholeNode } from '../../types/pihole';

export async function getPiholeNodes(): Promise<PiholeNode[]> {
	return apiFetch<PiholeNode[]>('/piholes');
}
