import { useState, FormEvent } from 'react';
import { PiholeNode } from '../../types/pihole';
import useInput from '../../hooks/useInput';
import useTextarea from '../../hooks/useTextarea';
import { formatFromNode, parsePiholeUrl } from '../../utils/urlUtils';
import { PiholeCreateBody, PiholePatchBody } from '../../lib/api/pihole';
import PasswordField from '../PasswordField';
import '../../styles/components/PiholeManagementList/pihole-node-form.scss';

type Mode = 'create' | 'edit';

interface PropsShared {
	mode: Mode;
	submitting: boolean;
	onCancel: () => void;
}

interface PropsCreate extends PropsShared {
	mode: 'create';
	node?: undefined;
	onSubmit: (node: PiholeCreateBody) => void | Promise<void>;
}

interface PropsEdit extends PropsShared {
	mode: 'edit';
	node: PiholeNode;
	onSubmit: (id: number, node: PiholePatchBody) => void | Promise<void>;
}

type Props = PropsCreate | PropsEdit;

export default function PiholeNodeForm(props: Props) {
	const { mode } = props;

	const initialUrl =
		mode === 'edit' ? formatFromNode(props.node.scheme, props.node.host, props.node.port) : '';

	const name = useInput(props.node?.name ?? '');
	const url = useInput(initialUrl);
	const [urlError, setUrlError] = useState<string>('');
	const description = useTextarea(props.node?.description ?? '');
	const password = useInput('');

	function testConnection() {
		try {
			// await testPiholeConnection();
		} catch (err: unknown) {
			console.error(err);
		}
	}

	function validateUrlCurrentValue() {
		try {
			parsePiholeUrl(url.value);
			setUrlError('');
			return true;
		} catch (err: unknown) {
			setUrlError((err as Error)?.message || 'Invalid URL');
			return false;
		}
	}

	function handleUrlBlur() {
		if (url.value.trim()) {
			validateUrlCurrentValue();
		}
	}

	function handleSubmit(e: FormEvent) {
		e.preventDefault();

		// Conversion
		if (!validateUrlCurrentValue()) {
			return;
		}
		const parsed = parsePiholeUrl(url.value);

		if (mode === 'create') {
			const newNode = {
				scheme: parsed.scheme,
				host: parsed.host,
				port: parsed.port,
				name: name.value,
				description: description.value,
				password: password.value,
			};
			return props.onSubmit(newNode);
		}

		const updatedNode = {
			scheme: parsed.scheme,
			host: parsed.host,
			port: parsed.port,
			name: name.value,
			description: description.value,
			password: password.value,
		};
		return props.onSubmit(props.node.id, updatedNode);
	}

	return (
		<form className='pihole-node-form' onSubmit={handleSubmit}>
			<label>
				Name
				<p className='hint'>
					Give the pihole node a short, descriptive name to help distinguish it from other
					instances
				</p>
				<input
					className='name-input'
					{...name}
					placeholder='(required) e.g. pihole1, etc.'
					disabled={props.submitting}
				/>
			</label>
			<label>
				Instance URL
				<input
					{...url}
					className='url-input'
					onBlur={handleUrlBlur}
					placeholder='e.g. pi.hole, 192.168.1.10:8080, https://host'
					aria-invalid={!!urlError}
					aria-describedby='url-error'
					disabled={props.submitting}
				/>
				{urlError && (
					<p id='url-error' className='error-text'>
						{urlError}
					</p>
				)}
			</label>
			<PasswordField
				label='Password'
				value={password.value}
				onChange={password.onChange}
				disabled={props.submitting}
				autoComplete='current-password'
			/>
			<label>
				Description
				<textarea
					{...description}
					placeholder='(optional) This is the first / primary node in the cluster, etc.'
					disabled={props.submitting}
				/>
			</label>
			<div className='button-bar'>
				<button
					type='button'
					onClick={testConnection}
					disabled={props.submitting}
					className='secondary'
				>
					Cancel
				</button>
				<button type='submit' disabled={props.submitting}>
					Save
				</button>
				<button
					type='button'
					onClick={props.onCancel}
					disabled={props.submitting}
					className='secondary'
				>
					Cancel
				</button>
			</div>
		</form>
	);
}
