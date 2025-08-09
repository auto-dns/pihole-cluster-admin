import { PiholeNode } from '../../types/pihole';
import useInput from '../../hooks/useInput';
import { HttpScheme } from '../../types';
import { FormEvent } from 'react';

interface Props {
	node?: PiholeNode;
	onSubmit: (node: PiholeNode, password: string) => void;
}

export default function PiholeNodeForm({ node, onSubmit }: Props) {
	const id = node?.id;
	const scheme = useInput(node?.scheme ?? '');
	const host = useInput(node?.host ?? '');
	const port = useInput(node?.port?.toString() ?? '');
	const name = useInput(node?.name ?? '');
	const description = useInput(node?.description ?? '');
	const password = useInput('');

	function handleSubmit(e: FormEvent) {
		e.preventDefault();

		// Conversion
		const portNumber = Number(port.value);

		const formNode = {
			host,
			port: portNumber,
			name,
			description,
		};
		onSubmit(formNode, password);
	}

	return (
		<form onSubmit={handleSubmit}>
			<label>
				Name
				<input {...name} />
			</label>
			<label>
				Scheme
				<input {...scheme} />
			</label>
			<label>
				Host
				<input {...host} />
			</label>
			<label>
				Port
				<input type='number' min={1} max={65535} {...port} />
			</label>
			<label>
				Password
				<input type='password' {...password} />
			</label>
			<label>
				Description
				<input {...description} />
			</label>
		</form>
	);
}
