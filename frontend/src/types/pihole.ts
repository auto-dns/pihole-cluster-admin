import { HttpScheme } from './';

export type PiholeNodeRef = {
	id: number;
	name: string;
	host: string;
};

export interface PiholeNode extends PiholeNodeRef {
	scheme: HttpScheme;
	port: number;
	description: string;
}
