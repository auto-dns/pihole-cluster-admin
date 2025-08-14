import classNames from 'classnames';
import React from 'react';
import styles from './StatusLight.module.scss';

type Mode = 'breathe' | 'blink';

export type StatusLightProps = {
	/** Accessible label (required for screen readers) */
	label: string;
	/** Tooltip/title */
	title?: string;

	/** Visuals */
	color?: string; // e.g. 'var(--accent-primary)' or '#10b981'
	size?: number; // px (default 10)
	ring?: boolean; // outer glow ring (default true)
	ringSize?: number; // px (default 3)
	ringOpacity?: number; // 0..1 (default 0.15)

	/** Animation */
	pulse?: boolean; // on/off (default false)
	durationMs?: number; // default 2400
	mode?: Mode; // 'breathe' | 'blink' (default 'breathe')
	breatheMinOpacity?: number; // 0..1, how dim it gets at mid-breath (default 0.75)

	className?: string;
	style?: React.CSSProperties;
	role?: React.AriaRole; // default 'status'
};

export default function StatusLight({
	label,
	title,
	color,
	size = 10,
	ring = true,
	ringSize = 3,
	ringOpacity = 0.15,
	pulse = false,
	durationMs = 2400,
	mode = 'breathe',
	breatheMinOpacity = 0.75,
	className,
	style,
	role = 'status',
}: StatusLightProps) {
	const cssVars = {
		'--light-color': color ?? 'var(--border-primary)',
		'--light-size': `${size}px`,
		'--ring-size': `${ringSize}px`,
		'--ring-opacity': String(ringOpacity),
		'--pulse-duration': `${durationMs}ms`,
		'--breathe-min-opacity': String(breatheMinOpacity),
	} as React.CSSProperties & Record<string, string>;

	// When animation parameters change, remount to restart the animation cleanly
	const animKey = `${pulse ? 1 : 0}-${mode}-${durationMs}-${breatheMinOpacity}-${color}-${ring ? 1 : 0}`;

	return (
		<span
			key={animKey}
			className={classNames(
				styles.light,
				pulse && styles.pulse,
				!ring && styles.noRing,
				className,
			)}
			data-mode={mode}
			aria-label={label}
			title={title}
			role={role}
			style={{ ...cssVars, ...style }}
		/>
	);
}
