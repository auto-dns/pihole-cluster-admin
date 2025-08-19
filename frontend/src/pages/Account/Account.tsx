import { FormEvent, useMemo, useState } from 'react';
import AppCard from '@/components/Layout/AppCard';
import PasswordField from '@/components/PasswordField/PasswordField';
import { useAuth } from '@/providers/AuthProvider';
import { validateUsername, validatePassword } from '@/utils/validation';
import styles from './Account.module.scss';
import { updateUser, updatePassword, UserPatchBody } from '../../lib/api/user';

export default function Account() {
	const { user, setUser, logout } = useAuth();
	const [username, setUsername] = useState(user?.username ?? '');
	const [unameBusy, setUnameBusy] = useState(false);
	const [pwdBusy, setPwdBusy] = useState(false);

	// username form state
	const [unameErr, setUnameErr] = useState<string>('');
	const [unameTouched, setUnameTouched] = useState(false);

	// password form state
	const [currentPwd, setCurrentPwd] = useState('');
	const [newPwd, setNewPwd] = useState('');
	const [confirmPwd, setConfirmPwd] = useState('');
	const [pwdSuccess, setPwdSuccess] = useState<string>('');
	const [pwdErr, setPwdErr] = useState<string>('');

	const created = useMemo(() => fmtDate(user?.createdAt), [user?.createdAt]);
	const updated = useMemo(() => fmtDate(user?.updatedAt), [user?.updatedAt]);

	function fmtDate(iso?: string) {
		if (!iso) return '—';
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return '—';
		return d.toLocaleString();
	}

	// ------- Change Username -------
	async function onSubmitUsername(e: FormEvent) {
		e.preventDefault();
		const result = validateUsername(username);
		setUnameErr(result.reason ?? '');
		if (!result.valid) return;
		try {
			setUnameBusy(true);
			const patch: UserPatchBody = { username: username.trim() };
			const updatedUser = await updateUser(user?.id ?? -1, patch);
			setUser(updatedUser);
			setUnameErr(''); // clear
		} catch (err: any) {
			setUnameErr(err?.message || 'Failed to change username');
		} finally {
			setUnameBusy(false);
		}
	}

	// ------- Change Password -------
	async function onSubmitPassword(e: FormEvent) {
		e.preventDefault();
		const result = validatePassword(newPwd, confirmPwd);
		setPwdErr(result.reason ?? '');
		if (!result.valid) return;
		try {
			setPwdBusy(true);
			const body = { currentPassword: currentPwd.trim(), newPassword: newPwd.trim() };
			await updatePassword(user?.id ?? -1, body);
			setPwdErr('');
			setCurrentPwd('');
			setNewPwd('');
			setConfirmPwd('');
			setPwdSuccess('Password successfully updated!');
		} catch (err: any) {
			setPwdErr(err?.message || 'Failed to change password');
		} finally {
			setPwdBusy(false);
		}
	}

	return (
		<div className={styles.accountPage}>
			<div className={styles.grid}>
				{/* User info */}
				<AppCard className={styles.card}>
					<h2 className={styles.cardTitle}>User Info</h2>
					<dl className={styles.kv}>
						<div>
							<dt>Username</dt>
							<dd>{user?.username ?? '—'}</dd>
						</div>
						<div>
							<dt>Created</dt>
							<dd>{created}</dd>
						</div>
						<div>
							<dt>Last Updated</dt>
							<dd>{updated}</dd>
						</div>
					</dl>
				</AppCard>

				{/* Change username */}
				<AppCard className={styles.card}>
					<h2 className={styles.cardTitle}>Change Username</h2>
					<form onSubmit={onSubmitUsername} className={styles.form}>
						<label>
							New username
							<input
								value={username}
								onChange={(e) => {
									setUsername(e.target.value);
									if (unameErr) {
										setUnameErr('');
									}
								}}
								onBlur={() => {
									setUnameTouched(true);
								}}
								aria-invalid={!!unameErr}
								aria-describedby='uname-error'
								placeholder='e.g. admin'
							/>
						</label>
						<p id='uname-error' className={styles.errorText}>
							{(unameTouched && unameErr) || '\u00A0'}
						</p>
						<div className={styles.actions}>
							<button type='submit' disabled={unameBusy}>
								{unameBusy ? 'Saving…' : 'Save'}
							</button>
						</div>
					</form>
				</AppCard>

				{/* Change password */}
				<AppCard className={styles.card}>
					<h2 className={styles.cardTitle}>Change Password</h2>
					<form onSubmit={onSubmitPassword} className={styles.form}>
						<PasswordField
							label='Current password'
							value={currentPwd}
							onChange={(e) => {
								setCurrentPwd(e.target.value);
								if (pwdErr) setPwdErr('');
								if (pwdSuccess) setPwdSuccess('');
							}}
							autoComplete='current-password'
						/>
						<PasswordField
							label='New password'
							value={newPwd}
							onChange={(e) => {
								setNewPwd(e.target.value);
								if (pwdErr) setPwdErr('');
								if (pwdSuccess) setPwdSuccess('');
							}}
							autoComplete='new-password'
						/>
						<PasswordField
							label='Confirm new password'
							value={confirmPwd}
							onChange={(e) => {
								setConfirmPwd(e.target.value);
								if (pwdErr) setPwdErr('');
								if (pwdSuccess) setPwdSuccess('');
							}}
							autoComplete='new-password'
						/>
						{pwdSuccess && !pwdErr ? (
							<p className={styles.successText}>{pwdSuccess || '\u00A0'}</p>
						) : (
							<p className={styles.errorText}>{pwdErr || '\u00A0'}</p>
						)}
						<div className={styles.actions}>
							<button type='submit' disabled={pwdBusy}>
								{pwdBusy ? 'Saving…' : 'Save'}
							</button>
						</div>
					</form>
				</AppCard>

				{/* Logout (kept simple) */}
				<AppCard className={styles.card}>
					<h2 className={styles.cardTitle}>Logout</h2>
					<div className={styles.actions}>
						<button className='danger' onClick={logout}>
							Logout
						</button>
					</div>
				</AppCard>
			</div>
		</div>
	);
}
