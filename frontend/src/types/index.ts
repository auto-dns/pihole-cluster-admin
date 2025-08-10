export type HttpScheme = 'http' | 'https';

export interface HttpError extends Error {
	status?: number;
}
