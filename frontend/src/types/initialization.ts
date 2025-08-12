export const PiholeInitStatus = {
	UNINITIALIZED: 'UNINITIALIZED',
	ADDED: 'ADDED',
	SKIPPED: 'SKIPPED',
} as const;

export type PiholeInitStatus = (typeof PiholeInitStatus)[keyof typeof PiholeInitStatus];

export interface FullInitStatus {
	userCreated: boolean;
	piholeStatus: PiholeInitStatus;
}
