import StatusLight from './StatusLight';
import { NodeHealth } from '../../types/health';

export default function PiholeStatusLight({
	name,
	health,
	fresh,
}: {
	name: string;
	health?: NodeHealth;
	fresh: boolean;
}) {
	// derive color/pulse from state
	let color = 'var(--border-primary)';
	let pulse = false;
	let durationMs = 2400;

	if (fresh && health) {
		switch (health.status) {
			case 'online':
				color = 'var(--accent-success)';
				pulse = true;
				durationMs = 2000;
				break;
			case 'degraded':
				color = 'var(--accent-warn)';
				pulse = true;
				durationMs = 1400;
				break;
			case 'offline':
				color = 'var(--accent-danger)';
				pulse = false;
				break;
		}
	}

	return (
		<StatusLight
			label={`${name} ${health?.status ?? 'stale'}`}
			title={`${name}: ${(health?.status ?? 'STALE').toUpperCase()} • ${health?.latencyMs ?? 0}ms`}
			color={color}
			pulse={pulse}
			durationMs={durationMs}
			size={10}
			mode='blink'
		/>
	);
}
