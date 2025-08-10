import { HttpScheme } from './';

export interface PiholeNode {
	id: number;
	scheme: HttpScheme;
	host: string;
	port: number;
	name: string;
	description: string;
}
