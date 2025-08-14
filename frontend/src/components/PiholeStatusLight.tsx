import classNames from 'classnames';
import { PiholeNode } from '../types/pihole';
import { NodeHealth, NodeStatus } from '../types/health';
import { useMemo } from 'react';
import styles from './PiholeStatusLight.module.scss';

type Props = {
	node: PiholeNode;
	health?: NodeHealth;
	fresh: boolean;
};

export default function PiholeStatusLight({ node, health, fresh }: Props) {
	type StatusClass = 'online' | 'degraded' | 'offline' | 'stale';
	const statusClass = useMemo<StatusClass>(() => {
		if (!fresh) return 'stale';
		if (health?.status === NodeStatus.ONLINE) return 'online';
		if (health?.status === NodeStatus.DEGRADED) return 'degraded';
		return 'offline';
	}, [fresh, health]);

	const aria = `${node.name} ${statusClass}`;
	const title = `${node.name}: ${health?.status?.toUpperCase?.() ?? 'UNKNOWN'} • ${health?.latencyMs ?? 0}ms`;
	const shouldPulse = fresh && (statusClass === 'online' || statusClass === 'degraded');

	return (
		<span
			className={classNames(styles.light, styles[statusClass], {
				[styles.pulse]: shouldPulse,
			})}
			aria-label={aria}
			title={title}
		/>
	);
}
