import { NavLink } from 'react-router';
import { ChevronRight, ChevronLeft, FileText, Home, List, SettingsIcon } from 'lucide-react';
import classNames from 'classnames';
import { useLayout } from '../../providers/LayoutProvider';
import { useClusterHealth } from '../../hooks/useClusterHealth';
import styles from './Sidebar.module.scss';
import { HealthSummary } from '../../types/health';
import StatusLight from '../StatusLight/StatusLight';

const links = [
	{ to: '/', label: 'Home', icon: Home, end: true },
	{ to: '/query', label: 'Query Log', icon: FileText },
	{ to: '/domains', label: 'Domains', icon: List },
	{ to: '/settings', label: 'Settings', icon: SettingsIcon },
];

export default function Sidebar() {
	const { isMobile, sidebarOpen: open, setSidebarOpen: setOpen } = useLayout();
	const { summary } = useClusterHealth();

	return (
		<>
			<aside
				className={classNames(styles.sidebar, {
					[styles.open]: open,
					[styles.collapsed]: !open,
				})}
			>
				<button
					className={styles.collapseButton}
					onClick={() => setOpen((v) => !v)}
					aria-label={open ? 'Collapse sidebar' : 'Expand sidebar'}
					title={open ? 'Collapse' : 'Expand'}
				>
					{open ? <ChevronLeft size={16} /> : <ChevronRight size={16} />}
				</button>
				<nav>
					{links.map(({ to, label, icon: Icon, end }) => (
						<NavLink
							key={to}
							to={to}
							end={end}
							className={({ isActive }) =>
								classNames(styles.navItem, { [styles.active]: isActive })
							}
							title={open ? label : undefined}
							aria-label={open ? label : undefined}
							onClick={() => {
								if (isMobile) setOpen(false);
							}}
						>
							<Icon size={18} className={styles.icon} />
							<span className={styles.label}>{label}</span>
						</NavLink>
					))}
				</nav>
				<Footer summary={summary} />
			</aside>
			{isMobile && (
				<div
					className={classNames(styles.backdrop, { [styles.show]: open })}
					onClick={() => setOpen(false)}
					aria-hidden='true'
				/>
			)}
		</>
	);
}

function Footer({ summary }: { summary: HealthSummary | undefined }) {
	const online = summary?.online ?? 0;
	const total = summary?.total ?? 0;

	let color;
	let pulse = false;
	let durationMs = 2400;

	if (total === 0) {
		color = 'var(--border-primary)';
	} else if (online === 0) {
		color = 'var(--accent-danger)';
	} else if (online != total) {
		color = 'var(--accent-warn)';
		pulse = true;
		durationMs = 1400;
	} else {
		color = 'var(--accent-success)';
		pulse = true;
	}

	return (
		<div className={styles.foot}>
			<div
				className={styles.clusterMini}
				data-count={`${online}/${total}`}
				title={`${online}/${total} nodes online`}
				aria-label={`${online} of ${total} nodes online`}
			>
				<StatusLight
					label={`${online} of ${total} nodes online`}
					title={`${online} of ${total} nodes online`}
					color={color}
					pulse={pulse}
					durationMs={durationMs}
					mode='blink'
				/>
				<strong>
					{online}/{total}
				</strong>{' '}
				<span className={styles.muted}>nodes</span>
			</div>
		</div>
	);
}
