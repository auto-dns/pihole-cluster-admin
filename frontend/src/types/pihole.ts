import { HttpScheme } from './';

export interface PiholeNode {
	id: number;
	scheme: HttpScheme;
	host: string;
	port: number;
	name: string;
	description: string;
}

// Transport types (request bodies)
type PiholeNodeDraft = Omit<PiholeNode, 'id'>;

export type PiholeCreateBody = PiholeNodeDraft & {
	password: string;
};

export type PiholePatchBody = Partial<PiholeNodeDraft> & {
	password?: string;
};
