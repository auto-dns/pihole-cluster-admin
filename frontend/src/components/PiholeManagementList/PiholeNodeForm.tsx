import { useState, FormEvent } from 'react';
import { PiholeNode } from '../../types/pihole';
import useInput from '../../hooks/useInput';
import useTextarea from '../../hooks/useTextarea';
import { formatFromNode, parsePiholeUrl } from '../../utils/urlUtils';
import { PiholeCreateBody, PiholePatchBody, testPiholeConnection } from '../../lib/api/pihole';
import PasswordField from '../PasswordField';
import { Check, XCircle, Loader2 } from 'lucide-react';
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

type TestState = 'idle' | 'pending' | 'success' | 'error';

export default function PiholeNodeForm(props: Props) {
	const { mode } = props;

	const initialUrl =
		mode === 'edit' ? formatFromNode(props.node.scheme, props.node.host, props.node.port) : '';

	const name = useInput(props.node?.name ?? '');
	const url = useInput(initialUrl);
	const [urlError, setUrlError] = useState<string>('');
	const description = useTextarea(props.node?.description ?? '');
	const password = useInput('');

	const [testState, setTestState] = useState<TestState>('idle');
	const [testMsg, setTestMsg] = useState<string>(''); // error or success text
	const [clearTimer, setClearTimer] = useState<number | null>(null);

	async function testConnection() {
		if (!validateUrlCurrentValue()) {
			return;
		}

		if (clearTimer) {
			window.clearTimeout(clearTimer);
			setClearTimer(null);
		}

		setTestState('pending');
		setTestMsg('Testing connection…');

		try {
			const { scheme, host, port } = parsePiholeUrl(url.value);
			await testPiholeConnection({
				scheme,
				host,
				port,
				password: password.value,
			});
			setTestState('success');
			setTestMsg('Connected successfully');
			const t = window.setTimeout(() => setTestState('idle'), 2500);
			setClearTimer(t);
		} catch (err: unknown) {
			console.error(err);
			setTestState('error');
			setTestMsg((err as Error)?.message || 'Connection failed');
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
				<div className='test-wrap'>
					<button
						type='button'
						onClick={testConnection}
						disabled={props.submitting || testState === 'pending'}
						className='secondary'
					>
						{testState === 'pending' ? (
							<>
								<Loader2 className='lucide spin' aria-hidden='true' />
								Testing…
							</>
						) : (
							'Test Connection'
						)}
					</button>

					<div
						className={`status-pill ${testState}`} // <-- fixed width pill
						role='status'
						aria-live='polite'
						aria-atomic='true'
					>
						{testState === 'success' && (
							<>
								<Check className='lucide' aria-hidden='true' />
								{testMsg || 'Connected successfully'}
							</>
						)}
						{testState === 'error' && (
							<>
								<XCircle className='lucide' aria-hidden='true' />
								{testMsg || 'Connection failed'}
							</>
						)}
						{testState === 'pending' && (
							<>
								<Loader2 className='lucide spin' aria-hidden='true' />
								{testMsg || 'Testing…'}
							</>
						)}
					</div>
				</div>
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
