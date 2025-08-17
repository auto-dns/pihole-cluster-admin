import classNames from 'classnames';
import React from 'react';
import styles from './StatusLight.module.scss';

type Mode = 'breathe' | 'blink';

export type StatusLightProps = {
	label: string;
	title?: string;

	// Visuals
	color?: string;
	size?: number;
	ring?: boolean;
	ringSize?: number;
	ringOpacity?: number;

	// Animation
	pulse?: boolean;
	durationMs?: number; // Total cycle length for breathe or blink (ms). e.g. 3000–5000 for blink
	mode?: Mode;
	breatheMinOpacity?: number; // Used by BOTH breathe & blink as the minimum (dim) opacity, e.g. 0.2–0.3 for visible blink

	className?: string;
	style?: React.CSSProperties;
	role?: React.AriaRole;
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
	durationMs = 3000,
	mode = 'breathe',
	breatheMinOpacity = 0.2,
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
		'--min-opacity': String(breatheMinOpacity),
	} as React.CSSProperties & Record<string, string>;

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
