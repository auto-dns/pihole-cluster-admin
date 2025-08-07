import { FullInitStatus } from '../types/initialization';

export function isFullyInitialized(status?: FullInitStatus): boolean {
	if (!status) return false;
	return status.userCreated && status.piholeStatus !== 'UNINITIALIZED';
}
