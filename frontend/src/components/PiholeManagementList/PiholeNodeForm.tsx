import { useEffect, useRef, useState, FormEvent } from 'react';
import { PiholeNode } from '../../types/pihole';
import useInput from '../../hooks/useInput';
import useTextarea from '../../hooks/useTextarea';
import { formatFromNode, parsePiholeUrl } from '../../utils/urlUtils';
import {
	PiholeCreateBody,
	PiholePatchBody,
	testPiholeInstanceConnection,
	testExistingPiholeConnection,
	PiholeTestExistingConnectionBody,
} from '../../lib/api/pihole';
import PasswordField from '../PasswordField/PasswordField';
import { Check, XCircle, Loader2 } from 'lucide-react';
import classNames from 'classnames';
import styles from './PiholeNodeForm.module.scss';

function ErrorText({ show, message }: { show: boolean; message: string }) {
	return <span className={styles.errorText}>{show ? message : '\u00A0'}</span>;
}

type Mode = 'create' | 'edit';

interface PropsShared {
	mode: Mode;
	submitting: boolean;
	onCancel: () => void;
	validateFormStatus: (
		name: string,
		url: string,
		password: string,
		description: string,
	) => boolean;
	processDirtyStatus?: (name: string, url: string, password: string, description: string) => void;
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
	onDelete?: (id: number) => void | Promise<void>;
	deleting?: boolean;
}

type Props = PropsCreate | PropsEdit;

type TestState = 'idle' | 'pending' | 'success' | 'error';

const DISPLAY_MS = 2500;
const FADE_MS = 500;

export default function PiholeNodeForm(props: Props) {
	// State
	const { mode } = props;

	const initialUrl =
		mode === 'edit' ? formatFromNode(props.node.scheme, props.node.host, props.node.port) : '';

	const name = useInput(props.node?.name ?? '');
	const url = useInput(initialUrl);
	const [urlError, setUrlError] = useState<string>('');
	const description = useTextarea(props.node?.description ?? '');
	const password = useInput('');

	const [formValid, setFormaValid] = useState<boolean>(false);

	// State - tests
	const [testState, setTestState] = useState<TestState>('idle');
	const [testMsg, setTestMsg] = useState<string>(''); // error or success text
	const [isFading, setIsFading] = useState(false);
	const displayTimer = useRef<number | null>(null);
	const fadeTimer = useRef<number | null>(null);
	// Allow save anyway if test fails
	const [allowSaveAnyway, setAllowSaveAnyway] = useState<boolean>(false);
	const testKey = `${url.value.trim()}|${password.value.trim()}}`;
	const [lastTestKey, setLastTestKey] = useState<string>('');
	const [lastTestOK, setLastTestOK] = useState<boolean>(false);

	// Effects
	useEffect(() => {
		return () => {
			// cleanup on unmount
			if (displayTimer.current) window.clearTimeout(displayTimer.current);
			if (fadeTimer.current) window.clearTimeout(fadeTimer.current);
		};
	}, []);

	useEffect(() => {
		props.processDirtyStatus?.(name.value, url.value, password.value, description.value);
		const valid = props.validateFormStatus(
			name.value,
			url.value,
			password.value,
			description.value,
		);
		setFormaValid(valid);
	}, [name.value, url.value, password.value, description.value]);

	// Misc functions
	function startFadeOut() {
		// show for DISPLAY_MS, then fade for FADE_MS, then hide (idle)
		if (displayTimer.current) window.clearTimeout(displayTimer.current);
		if (fadeTimer.current) window.clearTimeout(fadeTimer.current);

		displayTimer.current = window.setTimeout(() => {
			setIsFading(true);
			fadeTimer.current = window.setTimeout(() => {
				setIsFading(false);
				setTestState('idle');
			}, FADE_MS);
		}, DISPLAY_MS);
	}

	function buildEditOverrides(
		original: PiholeNode,
		urlStr: string,
		rawPassword: string,
	): PiholePatchBody {
		const { scheme, host, port } = parsePiholeUrl(urlStr);
		const patch: PiholeTestExistingConnectionBody = {};
		if (scheme !== original.scheme) patch.scheme = scheme;
		if (host !== original.host) patch.host = host;
		if (port !== original.port) patch.port = port;
		if (rawPassword.trim() !== '') patch.password = rawPassword;
		return patch;
	}

	async function testConnection() {
		if (displayTimer.current) window.clearTimeout(displayTimer.current);
		if (fadeTimer.current) window.clearTimeout(fadeTimer.current);
		setIsFading(false);

		if (!validateUrlCurrentValue()) return;
		setTestState('pending');
		setTestMsg('Testing connection…');

		try {
			if (mode === 'create') {
				const { scheme, host, port } = parsePiholeUrl(url.value);
				await testPiholeInstanceConnection({
					scheme,
					host,
					port,
					password: password.value,
				});
			} else {
				const overrides = buildEditOverrides(props.node, url.value, password.value);
				await testExistingPiholeConnection(props.node.id, overrides);
			}
			setTestState('success');
			setTestMsg('Connected successfully');
			setLastTestKey(testKey);
			setLastTestOK(true);
			setAllowSaveAnyway(false);
			startFadeOut();
		} catch (err: unknown) {
			console.error(err);
			setTestState('error');
			setTestMsg((err as Error)?.message || 'Connection failed');
			setLastTestKey(testKey);
			setLastTestOK(false);
			setAllowSaveAnyway(true);
			startFadeOut();
		}
	}

	// Validation
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

	// Event handling
	function handleUrlBlur() {
		if (url.value.trim()) validateUrlCurrentValue();
	}

	function handleCancel() {
		props.onCancel();
	}

	async function handleSubmit(e: FormEvent) {
		e.preventDefault();

		// Conversion
		if (!validateUrlCurrentValue()) return;

		const needsRetest = mode === 'create' || testKey !== lastTestKey || !lastTestOK;
		if (needsRetest && !allowSaveAnyway) {
			setIsFading(false);
			setTestState('pending');
			setTestMsg('Testing connection…');
			try {
				if (mode === 'create') {
					const { scheme, host, port } = parsePiholeUrl(url.value);
					await testPiholeInstanceConnection({
						scheme,
						host,
						port,
						password: password.value,
					});
				} else {
					const overrides = buildEditOverrides(props.node, url.value, password.value);
					await testExistingPiholeConnection(props.node.id, overrides);
				}
				setTestState('success');
				setTestMsg('Connected successfully');
				setLastTestKey(testKey);
				setLastTestOK(true);
				startFadeOut();
			} catch (err) {
				setTestState('error');
				setTestMsg((err as Error)?.message || 'Connection failed');
				setLastTestKey(testKey);
				setLastTestOK(false);
				setAllowSaveAnyway(true);
				return;
			}
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

	function handleDeleteClick() {
		if (props.mode !== 'edit' || !props.onDelete) return;
		void props.onDelete(props.node.id);
	}

	return (
		<form className={styles.form} onSubmit={handleSubmit}>
			<label>
				Name
				<p className={styles.hint}>
					Give the pihole node a short, descriptive name to help distinguish it from other
					instances
				</p>
				<input
					className={styles.nameInput}
					{...name}
					placeholder='(required) e.g. pihole1, etc.'
					disabled={props.submitting}
				/>
			</label>
			<label>
				Instance URL
				<input
					{...url}
					className={styles.urlInput}
					onBlur={handleUrlBlur}
					placeholder='e.g. pi.hole, 192.168.1.10:8080, https://host'
					aria-invalid={!!urlError}
					aria-describedby='url-error'
					disabled={props.submitting}
				/>
				<ErrorText show={!!urlError} message={urlError || ''} />
			</label>
			<PasswordField
				label='Password'
				className={styles.passwordInput}
				value={password.value}
				placeholder={mode === 'edit' ? 'Leave empty to use current' : 'Enter a password'}
				onChange={password.onChange}
				disabled={props.submitting}
				autoComplete='current-password'
			/>
			<div className={styles.testWrapper}>
				<button
					type='button'
					onClick={testConnection}
					disabled={props.submitting || testState === 'pending'}
					className='secondary'
				>
					{testState === 'pending' ? (
						<>
							<Loader2
								className={classNames(styles.lucide, styles.spin)}
								aria-hidden='true'
							/>
							Testing…
						</>
					) : (
						'Test Connection'
					)}
				</button>

				<div
					className={classNames(styles.statusPill, styles[testState], {
						[styles.fadeOut]: isFading,
					})}
					role='status'
					aria-live='polite'
					aria-atomic='true'
				>
					{testState === 'success' && (
						<>
							<Check className={styles.lucide} aria-hidden='true' />
							{testMsg || 'Connected successfully'}
						</>
					)}
					{testState === 'error' && (
						<>
							<XCircle className={styles.lucide} aria-hidden='true' />
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
			<label>
				Description
				<textarea
					{...description}
					placeholder='(optional) This is the first / primary node in the cluster, etc.'
					disabled={props.submitting}
				/>
			</label>
			<div className={styles.buttonBar}>
				{props.mode === 'edit' && props.onDelete && (
					<button
						type='button'
						onClick={handleDeleteClick}
						className='danger'
						disabled={props.submitting || props.deleting}
						title='Remove this node from the cluster'
					>
						{props.deleting ? (
							<>
								<Loader2
									className={classNames(styles.lucide, styles.spin)}
									aria-hidden='true'
								/>{' '}
								Deleting...
							</>
						) : (
							'Delete'
						)}
					</button>
				)}
				<button
					type='submit'
					title={allowSaveAnyway ? 'Will save without verifying connection' : undefined}
					disabled={props.submitting || !formValid}
				>
					{allowSaveAnyway ? 'Save Anyway' : 'Save'}
				</button>
				<button
					type='button'
					onClick={handleCancel}
					disabled={props.submitting}
					className='secondary'
				>
					Cancel
				</button>
			</div>
		</form>
	);
}
