import { NavLink } from 'react-router';
import { ChevronRight, ChevronLeft, FileText, Home, List, SettingsIcon } from 'lucide-react';
import classNames from 'classnames';
import { useLayout } from '../../providers/LayoutProvider';
import { useClusterHealth } from '../../hooks/useClusterHealth';
import styles from './Sidebar.module.scss';
import { HealthSummary } from '../../types/health';
import StatusLight from '../StatusLight/StatusLight';
import { Logo } from '../Logo/Logo';

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
				id='sidebar'
				className={classNames(styles.sidebar, { [styles.collapsed]: !open })}
				aria-label='Primary navigation'
			>
				{!open && !isMobile && (
					<button
						className={classNames(styles.toggleOpenButton, styles.closed)}
						onClick={() => setOpen(true)}
						aria-label='Expand sidebar'
						title='Expand'
					>
						<ChevronRight size={16} />
					</button>
				)}

				{open && !isMobile && (
					<div className={styles.header}>
						<div className={styles.headerGrid}>
							<div aria-hidden />
							{/* left spacer */}
							<div className={styles.brandTitle}>
								Pi-hole Cluster
								<br />
								Admin
							</div>
							<button
								className={styles.toggleButton}
								onClick={() => setOpen(false)}
								aria-label='Collapse sidebar'
								title='Collapse'
							>
								<ChevronLeft size={16} />
							</button>
						</div>

						<div className={styles.logoWrap} aria-hidden>
							<Logo size={144} />
						</div>
					</div>
				)}

				<nav className={styles.nav}>
					{links.map(({ to, label, icon: Icon, end }) => (
						<NavLink
							key={to}
							to={to}
							end={end}
							className={({ isActive }) =>
								classNames(styles.navItem, { [styles.active]: isActive })
							}
							title={!open ? label : undefined}
							aria-label={!open ? label : undefined}
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

			{/* Mobile backdrop */}
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

	let color: string;
	let pulse = false;
	let durationMs = 2400;

	if (total === 0) {
		color = 'var(--border-primary)';
	} else if (online === 0) {
		color = 'var(--accent-danger)';
	} else if (online !== total) {
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
