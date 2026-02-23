import type { LucideIcon } from 'lucide-react';
import { Shield, ShieldOff, ShieldHalf, AlertTriangle } from 'lucide-react';
import type { ClusterBlockingSummary } from '@/types/blocking';

export type BlockingDisplayVariant =
	| 'enabled'       // green shield – blocking
	| 'disabled'      // blue shield-off – not blocking (valid)
	| 'mixed'         // blue caution – partial valid state
	| 'mixed-errors'  // yellow caution – partial due to API errors
	| 'degraded';    // red caution – all nodes error

export type BlockingDisplayState = {
	icon: LucideIcon;
	/** CSS variable for icon color, e.g. var(--accent-success) */
	colorVar: string;
	variant: BlockingDisplayVariant;
	label: string;
};

const COLORS = {
	success: 'var(--accent-success)',
	primary: 'var(--accent-primary)',
	warn: 'var(--accent-warn)',
	danger: 'var(--accent-danger)',
} as const;

/**
 * Unified blocking status for cluster: icon, color, and label.
 * Used by the Blocking page and the sidebar cluster status badge.
 */
export function getBlockingDisplayState(
	summary: ClusterBlockingSummary | undefined | null,
): BlockingDisplayState {
	const mode = summary?.mode ?? 'degraded';
	const errors = summary?.counts?.errors ?? 0;
	const failed = summary?.counts?.failed ?? 0;
	const hasErrors = errors > 0 || failed > 0;

	if (mode === 'enabled') {
		return {
			icon: Shield,
			colorVar: COLORS.success,
			variant: 'enabled',
			label: 'Blocking',
		};
	}

	if (mode === 'disabled') {
		return {
			icon: ShieldOff,
			colorVar: COLORS.primary,
			variant: 'disabled',
			label: 'Not blocking',
		};
	}

	if (mode === 'mixed') {
		if (hasErrors) {
			return {
				icon: AlertTriangle,
				colorVar: COLORS.warn,
				variant: 'mixed-errors',
				label: 'Partial (errors)',
			};
		}
		return {
			icon: ShieldHalf,
			colorVar: COLORS.primary,
			variant: 'mixed',
			label: 'Partial',
		};
	}

	// degraded: all nodes failed / API errors
	return {
		icon: AlertTriangle,
		colorVar: COLORS.danger,
		variant: 'degraded',
		label: 'Degraded',
	};
}
