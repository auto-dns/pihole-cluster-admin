export type PiholeInitStatus = 'UNINITIALIZED' | 'ADDED' | 'SKIPPED';

export interface FullInitStatus {
	userCreated: boolean;
	piholeStatus: PiholeInitStatus;
}
