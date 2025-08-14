import { NavLink } from 'react-router';
import { ChevronRight, ChevronLeft, FileText, Home, List, SettingsIcon } from 'lucide-react';
import classNames from 'classnames';
import { useLocalStorageState } from '../../hooks/useLocalStorageState';
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
	const [collapsed, setCollapsed] = useLocalStorageState<boolean>(
		'pihole-cluster-admin.sidebarCollapsed',
		false,
		{ syncAcrossTabs: true },
	);
	const { summary } = useClusterHealth();

	return (
		<aside className={classNames(styles.sidebar, { [styles.collapsed]: collapsed })}>
			<button
				className={styles.collapseButton}
				onClick={() => setCollapsed((v) => !v)}
				aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
				title={collapsed ? 'Expand' : 'Collapse'}
			>
				{collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
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
						title={collapsed ? label : undefined}
						aria-label={collapsed ? label : undefined}
					>
						<Icon size={18} className='icon' />
						<span className={styles.label}>{label}</span>
					</NavLink>
				))}
			</nav>
			<Footer summary={summary} />
		</aside>
	);
}

function Footer({ summary }: { summary: HealthSummary | undefined }) {
	return (
		<div className={styles.foot}>
			<div
				className={styles.clusterMini}
				data-count={`${summary?.online}/${summary?.total}`}
				title={`${summary?.online}/${summary?.total} nodes online`}
				aria-label={`${summary?.online} of ${summary?.total} nodes online`}
			>
				<StatusLight label={`${summary?.online} of ${summary?.total} nodes online`} />
				<strong>
					{summary?.online}/{summary?.total}
				</strong>{' '}
				<span className={styles.muted}>nodes</span>
			</div>
		</div>
	);
}
