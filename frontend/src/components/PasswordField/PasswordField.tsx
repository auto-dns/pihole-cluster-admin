import { forwardRef, useId, useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import styles from './PasswordField.module.scss';
import classNames from 'classnames';

type PasswordFieldProps = {
	label?: string;
	id?: string;
	name?: string;
	value: string;
	onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
	onBlur?: (e: React.FocusEvent<HTMLInputElement>) => void;
	placeholder?: string;
	autoComplete?: 'current-password' | 'new-password';
	disabled?: boolean;
	error?: string;
	required?: boolean;
	className?: string; // lets you set max-width per page
};

const PasswordField = forwardRef<HTMLInputElement, PasswordFieldProps>(
	(
		{
			label = 'Password',
			id,
			name = 'password',
			value,
			onChange,
			onBlur,
			placeholder,
			autoComplete = 'current-password',
			disabled,
			error,
			required,
			className,
		},
		ref,
	) => {
		const reactId = useId();
		const inputId = id ?? `pw-${reactId}`;
		const errId = `${inputId}-error`;
		const [visible, setVisible] = useState(false);

		return (
			<div className={classNames(styles.pwField, className)}>
				<label htmlFor={inputId} className={styles.pwLabel}>
					{label}
					{required ? ' *' : ''}
				</label>

				<div className={styles.pwInputWrap}>
					<input
						ref={ref}
						id={inputId}
						name={name}
						type={visible ? 'text' : 'password'}
						value={value}
						onChange={onChange}
						onBlur={onBlur}
						placeholder={placeholder}
						autoComplete={autoComplete}
						disabled={disabled}
						aria-invalid={!!error}
						aria-describedby={error ? errId : undefined}
						className={styles.pwInput}
					/>
					<button
						type='button'
						className={styles.pwToggle}
						onClick={() => setVisible((v) => !v)}
						aria-label={visible ? 'Hide password' : 'Show password'}
						aria-pressed={visible}
						aria-controls={inputId}
						tabIndex={0}
						disabled={disabled}
					>
						{visible ? <EyeOff size={18} /> : <Eye size={18} />}
					</button>
				</div>

				{error && (
					<p id={errId} className='error-text'>
						{error}
					</p>
				)}
			</div>
		);
	},
);

PasswordField.displayName = 'PasswordField';
export default PasswordField;
